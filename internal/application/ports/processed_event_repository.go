package ports

import (
	"context"

	"github.com/be-users-metadata-service/internal/domain"
	"gorm.io/gorm"
)

// ProcessedEventRepository provides idempotency and processed event persistence.
type ProcessedEventRepository interface {
	Exists(ctx context.Context, eventID string) (bool, error)
	CreateTx(tx *gorm.DB, pe *domain.ProcessedEvent) error
}
