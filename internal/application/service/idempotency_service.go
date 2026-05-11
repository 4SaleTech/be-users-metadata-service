package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/be-users-metadata-service/internal/application/ports"
	"github.com/be-users-metadata-service/internal/domain"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// IdempotencyService checks and records event processing (with or without metadata update).
type IdempotencyService struct {
	txManager          ports.TransactionManager
	processedEventRepo ports.ProcessedEventRepository
	userRepo           ports.UserRepository
}

// NewIdempotencyService returns a new IdempotencyService.
func NewIdempotencyService(txManager ports.TransactionManager, processedEventRepo ports.ProcessedEventRepository, userRepo ports.UserRepository) *IdempotencyService {
	return &IdempotencyService{
		txManager:          txManager,
		processedEventRepo: processedEventRepo,
		userRepo:           userRepo,
	}
}

// Exists returns true if the event was already processed.
func (s *IdempotencyService) Exists(ctx context.Context, eventID string) (bool, error) {
	return s.processedEventRepo.Exists(ctx, eventID)
}

// RecordProcessedOnly records the event as processed without updating user metadata.
func (s *IdempotencyService) RecordProcessedOnly(ctx context.Context, eventID string, event *domain.Event, rawPayload []byte) error {
	pe := &domain.ProcessedEvent{
		EventID:     eventID,
		EventJSON:   buildProcessedEventJSON(event, rawPayload),
		ProcessedAt: time.Now().UTC(),
	}
	return s.txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		return s.processedEventRepo.CreateTx(tx, pe)
	})
}

// RecordSuccess updates user metadata (classified8) then records the event as processed (metadata DB). Two DBs, so no single transaction.
func (s *IdempotencyService) RecordSuccess(ctx context.Context, userID string, newMeta datatypes.JSON, eventID string, event *domain.Event, rawPayload []byte) error {
	if err := s.userRepo.UpdateMetaData(ctx, userID, newMeta); err != nil {
		return err
	}
	pe := &domain.ProcessedEvent{
		EventID:     eventID,
		EventJSON:   buildProcessedEventJSON(event, rawPayload),
		ProcessedAt: time.Now().UTC(),
	}
	return s.txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		return s.processedEventRepo.CreateTx(tx, pe)
	})
}

func buildProcessedEventJSON(event *domain.Event, rawPayload []byte) json.RawMessage {
	// Keep topic/routing_key only when they exist in publish-envelope payload.
	var envelope struct {
		Topic      string `json:"topic"`
		RoutingKey string `json:"routing_key"`
	}
	_ = json.Unmarshal(rawPayload, &envelope)

	var data interface{}
	if event != nil && len(event.Data) > 0 {
		_ = json.Unmarshal(event.Data, &data)
	}

	ts := time.Now().UTC().Format(time.RFC3339)
	if event != nil && !event.Timestamp.IsZero() {
		ts = event.Timestamp.UTC().Format(time.RFC3339)
	}

	out := struct {
		EventType  string      `json:"event_type"`
		UserID     string      `json:"user_id"`
		Timestamp  string      `json:"timestamp"`
		Topic      string      `json:"topic"`
		RoutingKey string      `json:"routing_key"`
		Data       interface{} `json:"data"`
	}{
		EventType:  "",
		UserID:     "",
		Timestamp:  ts,
		Topic:      envelope.Topic,
		RoutingKey: envelope.RoutingKey,
		Data:       data,
	}
	if event != nil {
		out.EventType = event.Type
		out.UserID = event.UserID
	}
	b, err := json.Marshal(out)
	if err != nil {
		return json.RawMessage(rawPayload)
	}
	return json.RawMessage(b)
}
