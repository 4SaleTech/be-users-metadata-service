package service

import (
	"context"
	"time"

	"github.com/be-users-metadata-service/internal/application/ports"
	"github.com/be-users-metadata-service/internal/domain"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// IdempotencyService checks and records event processing (with or without metadata update).
type IdempotencyService struct {
	txManager            ports.TransactionManager
	processedEventRepo   ports.ProcessedEventRepository
	userRepo             ports.UserRepository
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

// RecordSuccess updates user metadata and records the event as processed.
func (s *IdempotencyService) RecordSuccess(ctx context.Context, userID uuid.UUID, newMeta datatypes.JSON, eventID string, rawPayload []byte) error {
	pe := &domain.ProcessedEvent{
		EventID:     eventID,
		EventJSON:   rawPayload,
		ProcessedAt: time.Now().UTC(),
	}
	return s.txManager.WithTransaction(ctx, func(tx *gorm.DB) error {
		if err := s.userRepo.UpdateMetaDataTx(tx, userID, newMeta); err != nil {
			return err
		}
		return s.processedEventRepo.CreateTx(tx, pe)
	})
}