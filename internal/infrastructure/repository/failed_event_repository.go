package repository

import (
	"context"
	"time"

	"github.com/be-users-metadata-service/internal/domain"
	"github.com/be-users-metadata-service/internal/mappers"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FailedEventRepository handles failed_events table.
type FailedEventRepository struct {
	db *gorm.DB
}

// NewFailedEventRepository creates a new FailedEventRepository.
func NewFailedEventRepository(db *gorm.DB) *FailedEventRepository {
	return &FailedEventRepository{db: db}
}

// Create inserts a failed event record.
func (r *FailedEventRepository) Create(ctx context.Context, fe *domain.FailedEvent) error {
	row := mappers.FailedEventFromDomain(*fe)
	if row.ID == uuid.Nil {
		row.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(&row).Error
}

// CreateFromError builds and saves a FailedEvent from an error.
func (r *FailedEventRepository) CreateFromError(ctx context.Context, eventType string, payload []byte, errMsg string) error {
	now := time.Now()
	fe := &domain.FailedEvent{
		ID:           uuid.New(),
		EventType:    eventType,
		Payload:      payload,
		ErrorMessage: errMsg,
		CreatedAt:    now,
	}
	return r.Create(ctx, fe)
}
