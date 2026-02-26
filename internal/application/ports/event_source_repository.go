package ports

import (
	"context"

	"github.com/be-users-metadata-service/internal/domain"
)

// EventSourceRepository lists enabled event sources.
type EventSourceRepository interface {
	ListEnabled(ctx context.Context) ([]domain.EventSource, error)
}
