package ports

import (
	"context"

	"github.com/be-users-metadata-service/internal/domain"
)

// RuleRepository loads metadata rules from persistence.
type RuleRepository interface {
	FindByEventTypeAndVersion(ctx context.Context, eventType, eventVersion string) ([]domain.MetadataRule, error)
}
