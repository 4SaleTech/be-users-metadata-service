package bootstrap

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/be-users-metadata-service/internal/application/ports"
	"github.com/be-users-metadata-service/internal/application/service"
	"github.com/be-users-metadata-service/internal/application/usecase"
	"github.com/be-users-metadata-service/internal/config"
	"github.com/be-users-metadata-service/internal/infrastructure/database"
	"github.com/be-users-metadata-service/internal/infrastructure/logger"
	infraMessaging "github.com/be-users-metadata-service/internal/infrastructure/messaging"
	"github.com/be-users-metadata-service/internal/infrastructure/repository"
	consumerHandler "github.com/be-users-metadata-service/internal/interfaces/messaging"
	"gorm.io/gorm"
)

// App holds the wired application and runs the consumer.
type App struct {
	cfg            *config.Config
	log            *slog.Logger
	db             *gorm.DB // primary (service) DB
	usersDB        *gorm.DB // users DB (clas_users); may be same as db when single-DB)
	consumer       *infraMessaging.Consumer
	streamConsumer *infraMessaging.SuperStreamConsumer
	handler        *consumerHandler.RabbitMQConsumer
	eventSrc       ports.EventSourceRepository
	wg             sync.WaitGroup
}

// New builds and wires the application.
func New(cfg *config.Config) (*App, error) {
	log := logger.New(logger.LevelFromString(cfg.Log.Level))

	db, err := setupDB(cfg, log)
	if err != nil {
		return nil, err
	}

	usersDB, err := setupUsersDB(cfg, log)
	if err != nil {
		return nil, err
	}
	if usersDB == nil {
		usersDB = db
		log.Info("using primary DB for users table (single-DB mode)")
	} else {
		log.Info("users DB connected (two-DB mode)", "table", cfg.UsersDB.TableName)
	}

	consumer, err := infraMessaging.NewConsumer(cfg.RabbitMQ.URL, cfg.RabbitMQ.Prefetch)
	if err != nil {
		log.Error("rabbitmq init failed", "error", err)
		return nil, err
	}

	txManager := database.NewGormTransactionManager(db)
	ruleRepo := repository.NewMetadataRuleRepository(db)
	userRepo := repository.NewUserRepository(usersDB, cfg.UsersDB.TableName)
	processedEventRepo := repository.NewProcessedEventRepository(db)
	failedEventRepo := repository.NewFailedEventRepository(db)
	eventSourceRepo := repository.NewEventSourceRepository(db)

	ruleEngine := service.NewRuleEngine()
	metadataExecutor := service.NewMetadataExecutor()
	idempotency := service.NewIdempotencyService(txManager, processedEventRepo, userRepo)

	processEvent := usecase.NewProcessEvent(ruleRepo, userRepo, failedEventRepo, ruleEngine, metadataExecutor, idempotency, log)
	handler := consumerHandler.NewRabbitMQConsumer(processEvent, failedEventRepo, log)

	app := &App{
		cfg:      cfg,
		log:      log,
		db:       db,
		usersDB:  usersDB,
		consumer: consumer,
		handler:  handler,
		eventSrc: eventSourceRepo,
	}

	if cfg.RabbitMQ.StreamEnabled && cfg.RabbitMQ.SuperStreamName != "" {
		streamCfg := infraMessaging.StreamConsumerConfig{
			Host:     cfg.RabbitMQ.StreamHost,
			Port:     cfg.RabbitMQ.StreamPort,
			User:     cfg.RabbitMQ.User,
			Password: cfg.RabbitMQ.Password,
		}
		streamConsumer, err := infraMessaging.NewSuperStreamConsumer(
			streamCfg,
			cfg.RabbitMQ.SuperStreamName,
			"user-metadata-service",
			handler.Handle,
			log,
		)
		if err != nil {
			log.Error("super stream consumer init failed", "error", err)
			_ = consumer.Close()
			return nil, err
		}
		app.streamConsumer = streamConsumer
	}

	return app, nil
}

func setupDB(cfg *config.Config, log *slog.Logger) (*gorm.DB, error) {
	dbCfg := database.Config{
		DSN:             cfg.DB.DSN,
		MaxOpenConns:    cfg.DB.MaxOpenConns,
		MaxIdleConns:    cfg.DB.MaxIdleConns,
		ConnMaxLifetime: cfg.DB.ConnMaxLifetime,
	}
	db, err := database.NewDB(dbCfg)
	if err != nil {
		log.Error("database init failed", "error", err)
		return nil, err
	}
	if err := database.Ping(context.Background(), db); err != nil {
		log.Error("database ping failed", "error", err)
		return nil, err
	}
	log.Info("database connected")
	return db, nil
}

func setupUsersDB(cfg *config.Config, log *slog.Logger) (*gorm.DB, error) {
	if cfg.UsersDB.DSN == "" {
		return nil, nil
	}
	usersDB, err := database.NewUsersDB(struct {
		DSN             string
		MaxOpenConns    int
		MaxIdleConns    int
		ConnMaxLifetime time.Duration
	}{
		DSN:             cfg.UsersDB.DSN,
		MaxOpenConns:    cfg.UsersDB.MaxOpenConns,
		MaxIdleConns:    cfg.UsersDB.MaxIdleConns,
		ConnMaxLifetime: cfg.UsersDB.ConnMaxLifetime,
	})
	if err != nil {
		log.Error("users database init failed", "error", err)
		return nil, err
	}
	if err := database.Ping(context.Background(), usersDB); err != nil {
		log.Error("users database ping failed", "error", err)
		return nil, err
	}
	return usersDB, nil
}

// Run starts consumer workers and blocks until shutdown.
func (a *App) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := a.runConsumerWorkers(ctx); err != nil {
		return err
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	a.log.Info("shutdown signal received")
	cancel()
	a.wg.Wait()
	_ = a.consumer.Close()
	if a.streamConsumer != nil {
		_ = a.streamConsumer.Close()
	}
	a.log.Info("shutdown complete")
	return nil
}

func (a *App) runConsumerWorkers(ctx context.Context) error {
	if a.streamConsumer != nil {
		a.wg.Add(1)
		go func() {
			defer a.wg.Done()
			a.log.Info("super stream consumer started", "stream", a.cfg.RabbitMQ.SuperStreamName)
			// SuperStreamConsumer runs until Close(); block until ctx is done then close.
			<-ctx.Done()
			_ = a.streamConsumer.Close()
		}()
	}

	sources, err := a.eventSrc.ListEnabled(ctx)
	if err != nil {
		return err
	}
	consumeHandler := a.makeAMQPHandler()
	for _, src := range sources {
		queueName := "user_metadata_" + sanitizeQueueName(src.TopicName)
		if err := a.consumer.EnsureQueue(src.TopicName, queueName, "#"); err != nil {
			a.log.Warn("ensure queue failed", "topic", src.TopicName, "error", err)
			continue
		}
		a.wg.Add(1)
		go func(qName string) {
			defer a.wg.Done()
			if err := a.consumer.Consume(ctx, qName, consumeHandler); err != nil && ctx.Err() == nil {
				a.log.Error("consume failed", "queue", qName, "error", err)
			}
		}(queueName)
	}
	a.log.Info("user-metadata-service started", "sources", len(sources))
	return nil
}

func (a *App) makeAMQPHandler() func(context.Context, []byte, amqp.Delivery) {
	return func(ctx context.Context, body []byte, d amqp.Delivery) {
		start := time.Now()
		eventLog := a.log
		if eventID, ok := parseEventID(body); ok && eventID != "" {
			eventLog = a.log.With("event_id", eventID)
		}
		err := a.handler.Handle(ctx, body)
		if err != nil {
			eventLog.Error("event handling failed, nacking", "error", err, "duration_ms", time.Since(start).Milliseconds())
			_ = d.Nack(false, true)
			return
		}
		eventLog.Info("event handled", "duration_ms", time.Since(start).Milliseconds())
		_ = d.Ack(false)
	}
}

func parseEventID(body []byte) (string, bool) {
	var m struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(body, &m); err != nil {
		return "", false
	}
	return m.ID, m.ID != ""
}

func sanitizeQueueName(s string) string {
	var b []byte
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			b = append(b, byte(r))
		} else if r >= 'A' && r <= 'Z' {
			b = append(b, byte(r)+32)
		}
	}
	if len(b) == 0 {
		return "default"
	}
	return string(b)
}
