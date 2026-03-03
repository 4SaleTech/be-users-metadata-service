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
	streamlogs "github.com/rabbitmq/rabbitmq-stream-go-client/pkg/logs"

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
	cfg             *config.Config
	log             *slog.Logger
	db              *gorm.DB
	usersDB         *gorm.DB // classified8 (clas_users)
	consumer        *infraMessaging.Consumer
	streamConsumers []*infraMessaging.SuperStreamConsumer
	handler         *consumerHandler.RabbitMQConsumer
	eventSrc        ports.EventSourceRepository
	wg              sync.WaitGroup
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

	consumer, err := infraMessaging.NewConsumer(cfg.RabbitMQ.URL, cfg.RabbitMQ.Prefetch)
	if err != nil {
		log.Error("rabbitmq init failed", "error", err)
		return nil, err
	}

	txManager := database.NewGormTransactionManager(db)
	ruleRepo := repository.NewMetadataRuleRepository(db)
	userRepo := repository.NewUserRepository(usersDB)
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
	dbCfg := database.Config{
		DSN:             cfg.UsersDB.DSN,
		MaxOpenConns:    cfg.UsersDB.MaxOpenConns,
		MaxIdleConns:    cfg.UsersDB.MaxIdleConns,
		ConnMaxLifetime: cfg.UsersDB.ConnMaxLifetime,
	}
	db, err := database.NewDBNoMigrate(dbCfg)
	if err != nil {
		log.Error("users database init failed", "error", err)
		return nil, err
	}
	if err := database.Ping(context.Background(), db); err != nil {
		log.Error("users database ping failed", "error", err)
		return nil, err
	}
	log.Info("users database (classified8) connected")
	return db, nil
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
	for _, sc := range a.streamConsumers {
		_ = sc.Close()
	}
	a.log.Info("shutdown complete")
	return nil
}

func (a *App) runConsumerWorkers(ctx context.Context) error {
	sources, err := a.eventSrc.ListEnabled(ctx)
	if err != nil {
		return err
	}

	// Super stream consumers: one per event source (stream name = topic_name from event_sources)
	if a.cfg.RabbitMQ.StreamEnabled {
		streamlogs.LogLevel = streamlogs.DEBUG
		streamCfg := infraMessaging.StreamConsumerConfig{
			Host:     a.cfg.RabbitMQ.Host,
			Port:     a.cfg.RabbitMQ.StreamPort,
			User:     a.cfg.RabbitMQ.User,
			Password: a.cfg.RabbitMQ.Password,
		}
		for _, src := range sources {
			sc, err := infraMessaging.NewSuperStreamConsumer(
				streamCfg,
				src.TopicName,
				"user-metadata-service",
				a.handler.Handle,
				a.log,
			)
			if err != nil {
				a.log.Warn("super stream consumer init failed", "stream", src.TopicName, "error", err)
				continue
			}
			a.streamConsumers = append(a.streamConsumers, sc)
			a.wg.Add(1)
			streamName := src.TopicName
			consumerToClose := sc
			go func() {
				defer a.wg.Done()
				a.log.Info("super stream consumer started", "stream", streamName)
				<-ctx.Done()
				_ = consumerToClose.Close()
			}()
		}
	}

	// AMQP consumers: only when stream is disabled (stream and AMQP both use event_sources; same name can be a super stream with direct exchange, so we don't declare topic exchange when stream enabled)
	if !a.cfg.RabbitMQ.StreamEnabled {
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
	}
	a.log.Info("user-metadata-service started", "sources", len(sources), "stream_enabled", a.cfg.RabbitMQ.StreamEnabled)
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
