package repository

import (
	"context"

	"github.com/be-users-metadata-service/internal/domain"
	"github.com/be-users-metadata-service/internal/infrastructure/entity"
	"github.com/be-users-metadata-service/internal/mappers"
	"gorm.io/gorm"
)

// MetadataRuleRepository provides access to metadata_rules and metadata_rule_actions.
type MetadataRuleRepository struct {
	db *gorm.DB
}

// NewMetadataRuleRepository creates a new MetadataRuleRepository.
func NewMetadataRuleRepository(db *gorm.DB) *MetadataRuleRepository {
	return &MetadataRuleRepository{db: db}
}

// FindByEventTypeAndVersion returns enabled rules for the given event type (and optional version), ordered by priority desc.
func (r *MetadataRuleRepository) FindByEventTypeAndVersion(ctx context.Context, eventType, eventVersion string) ([]domain.MetadataRule, error) {
	var rows []entity.MetadataRule
	q := r.db.WithContext(ctx).Where("event_type = ? AND enabled = ?", eventType, true)
	if eventVersion != "" {
		q = q.Where("(event_version = '' OR event_version = ?)", eventVersion)
	} else {
		q = q.Where("(event_version = '' OR event_version IS NULL)")
	}
	if err := q.Preload("Actions").Order("priority DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.MetadataRule, len(rows))
	for idx := range rows {
		out[idx] = mappers.MetadataRuleToDomain(rows[idx])
	}
	return out, nil
}

// ListAllEnabled returns all enabled rules with actions (for cache warming).
func (r *MetadataRuleRepository) ListAllEnabled(ctx context.Context) ([]domain.MetadataRule, error) {
	var rows []entity.MetadataRule
	if err := r.db.WithContext(ctx).Where("enabled = ?", true).Preload("Actions").Order("priority DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]domain.MetadataRule, len(rows))
	for idx := range rows {
		out[idx] = mappers.MetadataRuleToDomain(rows[idx])
	}
	return out, nil
}
