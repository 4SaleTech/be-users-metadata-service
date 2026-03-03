package messaging

import (
	"context"
	"log/slog"
	"math/rand"
	"strings"
	"sync"
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
// Uses Single Active Consumer with explicit QueryOffset per partition for reliable resume across restarts.
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

	log = log.With("stream", superStreamName)

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

	sac := stream.NewSingleActiveConsumer(
		func(partition string, isActive bool) stream.OffsetSpecification {
			offset, err := env.QueryOffset(consumerName, partition)
			if err != nil {
				log.Info("consumer promoted, starting from first (no stored offset)", "partition", partition, "is_active", isActive)
				return stream.OffsetSpecification{}.First()
			}
			restart := stream.OffsetSpecification{}.Offset(offset + 1)
			log.Info("consumer promoted, resuming from stored offset", "partition", partition, "offset", offset+1, "is_active", isActive)
			return restart
		},
	)

	superConsumer, err := env.NewSuperStreamConsumer(superStreamName, handleMessages,
		stream.NewSuperStreamConsumerOptions().
			SetConsumerName(consumerName).
			SetSingleActiveConsumer(sac).
			SetOffset(stream.OffsetSpecification{}.First())) // ignored when SAC active
	if err != nil {
		_ = env.Close()
		return nil, err
	}

	go handlePartitionClose(superConsumer.NotifyPartitionClose(1), env, consumerName, log)
	return &SuperStreamConsumer{
		env:             env,
		consumer:        superConsumer,
		handler:         handler,
		log:             log,
		superStreamName: superStreamName,
		consumerName:    consumerName,
	}, nil
}

var reconnectMu sync.Mutex // serialize reconnects to avoid cascade when multiple partitions close at once

func handlePartitionClose(ch <-chan stream.CPartitionClose, env *stream.Environment, consumerName string, log *slog.Logger) {
	lastReconnect := make(map[string]time.Time)
	const minReconnectInterval = 45 * time.Second
	for ev := range ch {
		r := ev.Event.Reason
		if !strings.EqualFold(r, stream.SocketClosed) && !strings.EqualFold(r, stream.MetaDataUpdate) && !strings.EqualFold(r, stream.ZombieConsumer) {
			continue
		}
		// Serialize and rate-limit: only one reconnect at a time, skip if this partition was just reconnected
		reconnectMu.Lock()
		if last, ok := lastReconnect[ev.Partition]; ok && time.Since(last) < minReconnectInterval {
			reconnectMu.Unlock()
			log.Debug("partition close skipped (recent reconnect)", "partition", ev.Partition)
			continue
		}
		reconnectMu.Unlock()

		sleepSec := rand.Intn(15) + 10 // 10–25s backoff to let network settle (was 2–7s)
		log.Warn("partition closed unexpectedly, reconnecting", "partition", ev.Partition, "reason", ev.Event.Reason, "sleep_sec", sleepSec)
		time.Sleep(time.Duration(sleepSec) * time.Second)

		reconnectMu.Lock()
		restart := stream.OffsetSpecification{}.First()
		if offset, err := env.QueryOffset(consumerName, ev.Partition); err == nil {
			restart = stream.OffsetSpecification{}.Offset(offset + 1)
		}
		err := ev.Context.ConnectPartition(ev.Partition, restart)
		if err != nil {
			reconnectMu.Unlock()
			log.Error("partition reconnect failed", "partition", ev.Partition, "error", err)
			continue
		}
		lastReconnect[ev.Partition] = time.Now()
		reconnectMu.Unlock()
		log.Info("partition reconnected", "partition", ev.Partition)
	}
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
