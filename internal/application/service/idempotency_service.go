package service

import (
	"context"
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
func (s *IdempotencyService) RecordProcessedOnly(ctx context.Context, eventID string, rawPayload []byte) error {
	pe := &domain.ProcessedEvent{
		EventID:     eventID,
		EventJSON:   rawPayload,
		ProcessedAt: time.Now().UTC(),
	}
	return s.txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		return s.processedEventRepo.CreateTx(tx, pe)
	})
}

// RecordSuccess updates user metadata (classified8) then records the event as processed (metadata DB). Two DBs, so no single transaction.
func (s *IdempotencyService) RecordSuccess(ctx context.Context, userID string, newMeta datatypes.JSON, eventID string, rawPayload []byte) error {
	if err := s.userRepo.UpdateMetaData(ctx, userID, newMeta); err != nil {
		return err
	}
	pe := &domain.ProcessedEvent{
		EventID:     eventID,
		EventJSON:   rawPayload,
		ProcessedAt: time.Now().UTC(),
	}
	return s.txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		return s.processedEventRepo.CreateTx(tx, pe)
	})
}
