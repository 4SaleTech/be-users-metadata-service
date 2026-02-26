package ports

import "context"

// FailedEventRepository persists failed events.
type FailedEventRepository interface {
	CreateFromError(ctx context.Context, eventType string, payload []byte, errMsg string) error
}
