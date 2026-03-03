package messaging

import (
	"context"
	"log/slog"
	"time"

	streamamqp "github.com/rabbitmq/rabbitmq-stream-go-client/pkg/amqp"
	"github.com/rabbitmq/rabbitmq-stream-go-client/pkg/stream"
)

// StreamConsumerConfig holds connection params for the stream protocol (port 5552).
type StreamConsumerConfig struct {
	Host     string
	Port     int
	User     string
	Password string
}

// SuperStreamConsumer consumes from a RabbitMQ super stream and calls handler for each message body.
type SuperStreamConsumer struct {
	env             *stream.Environment
	consumer        *stream.SuperStreamConsumer
	handler         func(context.Context, []byte) error
	log             *slog.Logger
	superStreamName string
	consumerName    string
}

// NewSuperStreamConsumer creates a stream environment and a super stream consumer (not started).
// When the broker advertises internal hostnames (e.g. in Kubernetes), use AddressResolver so the client connects to cfg.Host:cfg.Port instead.
func NewSuperStreamConsumer(cfg StreamConsumerConfig, superStreamName, consumerName string, handler func(context.Context, []byte) error, log *slog.Logger) (*SuperStreamConsumer, error) {
	resolver := stream.AddressResolver{Host: cfg.Host, Port: cfg.Port}
	env, err := stream.NewEnvironment(
		stream.NewEnvironmentOptions().
			SetHost(cfg.Host).
			SetPort(cfg.Port).
			SetUser(cfg.User).
			SetPassword(cfg.Password).
			SetAddressResolver(resolver),
	)
	if err != nil {
		return nil, err
	}

	handleMessages := func(consumerContext stream.ConsumerContext, message *streamamqp.Message) {
		body := message.GetData()
		if len(body) == 0 {
			log.Debug("stream message with empty body, skipping")
			_ = consumerContext.Consumer.StoreOffset()
			return
		}
		ctx := context.Background()
		start := time.Now()
		if err := handler(ctx, body); err != nil {
			log.Error("stream message handling failed, offset not stored", "error", err, "duration_ms", time.Since(start).Milliseconds())
			return
		}
		if err := consumerContext.Consumer.StoreOffset(); err != nil {
			log.Warn("store offset failed", "error", err)
			return
		}
		log.Debug("stream message handled", "duration_ms", time.Since(start).Milliseconds())
	}

	// LastConsumed: resume from stored offset after deploy/restart; if no offset stored, start from first.
	superConsumer, err := env.NewSuperStreamConsumer(superStreamName, handleMessages,
		stream.NewSuperStreamConsumerOptions().
			SetConsumerName(consumerName).
			SetOffset(stream.OffsetSpecification{}.LastConsumed()))
	if err != nil {
		_ = env.Close()
		return nil, err
	}

	return &SuperStreamConsumer{
		env:             env,
		consumer:        superConsumer,
		handler:         handler,
		log:             log,
		superStreamName: superStreamName,
		consumerName:    consumerName,
	}, nil
}

// Close closes the super stream consumer and the environment.
func (s *SuperStreamConsumer) Close() error {
	if s.consumer != nil {
		_ = s.consumer.Close()
		s.consumer = nil
	}
	if s.env != nil {
		err := s.env.Close()
		s.env = nil
		return err
	}
	return nil
}
