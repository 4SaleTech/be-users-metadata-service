package repository

import (
	"context"

	"github.com/be-users-metadata-service/internal/domain"
	"github.com/be-users-metadata-service/internal/infrastructure/entity"
	"github.com/be-users-metadata-service/internal/mappers"
	"gorm.io/gorm"
)

// ProcessedEventRepository handles idempotency via processed_events.
type ProcessedEventRepository struct {
	db *gorm.DB
}

// NewProcessedEventRepository creates a new ProcessedEventRepository.
func NewProcessedEventRepository(db *gorm.DB) *ProcessedEventRepository {
	return &ProcessedEventRepository{db: db}
}

// Exists returns true if event_id was already processed.
func (r *ProcessedEventRepository) Exists(ctx context.Context, eventID string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.ProcessedEvent{}).Where("event_id = ?", eventID).Count(&count).Error
	return count > 0, err
}

// Create inserts a processed event (call within transaction).
func (r *ProcessedEventRepository) Create(ctx context.Context, pe *domain.ProcessedEvent) error {
	row := mappers.ProcessedEventFromDomain(*pe)
	return r.db.WithContext(ctx).Create(&row).Error
}

// CreateTx inserts a processed event using the given transaction.
func (r *ProcessedEventRepository) CreateTx(tx *gorm.DB, pe *domain.ProcessedEvent) error {
	row := mappers.ProcessedEventFromDomain(*pe)
	return tx.Create(&row).Error
}
