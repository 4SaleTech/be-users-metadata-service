package database

import (
	"context"

	"github.com/be-users-metadata-service/internal/application/ports"
	"gorm.io/gorm"
)

// GormTransactionManager implements ports.TransactionManager using gorm.DB.
type GormTransactionManager struct {
	db *gorm.DB
}

// NewGormTransactionManager returns a new GormTransactionManager.
func NewGormTransactionManager(db *gorm.DB) ports.TransactionManager {
	return &GormTransactionManager{db: db}
}

// WithTransaction runs fn inside a transaction.
func (g *GormTransactionManager) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return g.db.WithContext(ctx).Transaction(fn)
}
