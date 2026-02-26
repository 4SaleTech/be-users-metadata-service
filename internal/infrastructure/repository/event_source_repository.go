package repository

import (
	"context"

	"github.com/be-users-metadata-service/internal/domain"
	"github.com/be-users-metadata-service/internal/infrastructure/entity"
	"github.com/be-users-metadata-service/internal/mappers"
	"gorm.io/gorm"
)

// EventSourceRepository provides access to event_sources.
type EventSourceRepository struct {
	db *gorm.DB
}

// NewEventSourceRepository creates a new EventSourceRepository.
func NewEventSourceRepository(db *gorm.DB) *EventSourceRepository {
	return &EventSourceRepository{db: db}
}

// ListEnabled returns all enabled event sources.
func (r *EventSourceRepository) ListEnabled(ctx context.Context) ([]domain.EventSource, error) {
	var rows []entity.EventSource
	if err := r.db.WithContext(ctx).Where("enabled = ?", true).Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.EventSource, len(rows))
	for idx := range rows {
		out[idx] = mappers.EventSourceToDomain(rows[idx])
	}
	return out, nil
}
