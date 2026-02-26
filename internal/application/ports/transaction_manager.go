package ports

import (
	"context"

	"gorm.io/gorm"
)

// TransactionManager runs database transactions.
type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error
}
