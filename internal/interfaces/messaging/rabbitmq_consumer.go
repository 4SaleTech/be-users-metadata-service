package messaging

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/be-users-metadata-service/internal/application/ports"
	"github.com/be-users-metadata-service/internal/application/usecase"
	"github.com/be-users-metadata-service/internal/domain"
)

// RabbitMQConsumer handles incoming AMQP messages: parses to domain.Event and calls the process-event use case.
type RabbitMQConsumer struct {
	processEvent    *usecase.ProcessEvent
	failedEventRepo ports.FailedEventRepository
	log             *slog.Logger
}

// NewRabbitMQConsumer returns a new RabbitMQConsumer.
func NewRabbitMQConsumer(processEvent *usecase.ProcessEvent, failedEventRepo ports.FailedEventRepository, log *slog.Logger) *RabbitMQConsumer {
	return &RabbitMQConsumer{
		processEvent:    processEvent,
		failedEventRepo: failedEventRepo,
		log:             log,
	}
}

// Handle parses body into domain.Event and runs ProcessEvent. On parse failure, records failed event and returns nil (ack).
func (h *RabbitMQConsumer) Handle(ctx context.Context, body []byte) error {
	h.log.Debug("handling message", "body_size", len(body))
	var event domain.Event
	if err := json.Unmarshal(body, &event); err != nil {
		h.log.Warn("parse failed, recording failed event", "error", err.Error())
		eventType := eventTypeFromPayload(body)
		_ = h.failedEventRepo.CreateFromError(ctx, eventType, body, "invalid json: "+err.Error())
		return nil
	}
	h.log.Debug("message parsed", "event_type", event.Type, "event_id", event.EventID())
	if err := h.processEvent.ProcessEvent(ctx, event, body); err != nil {
		h.log.Error("process event failed", "event_id", event.EventID(), "event_type", event.Type, "error", err)
		return err
	}
	return nil
}

// eventTypeFromPayload tries to read event_type or type from raw JSON so failed_events.event_type is set correctly.
func eventTypeFromPayload(body []byte) string {
	var raw struct {
		Type      string `json:"type"`
		EventType string `json:"event_type"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return "unknown"
	}
	if raw.EventType != "" {
		return raw.EventType
	}
	if raw.Type != "" {
		return raw.Type
	}
	return "unknown"
}
