package database

import (
	"context"
	"time"

	"github.com/be-users-metadata-service/internal/infrastructure/entity"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Config holds DB connection settings.
type Config struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// NewDB creates a GORM DB and runs AutoMigrate for service tables (not clas_users).
func NewDB(cfg Config) (*gorm.DB, error) {
	db, err := openDB(cfg)
	if err != nil {
		return nil, err
	}
	if err := autoMigrate(db); err != nil {
		return nil, err
	}
	return db, nil
}

// NewDBNoMigrate creates a GORM DB without running migrations. Use for the classified8 DB (clas_users).
func NewDBNoMigrate(cfg Config) (*gorm.DB, error) {
	return openDB(cfg)
}

func openDB(cfg Config) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	if cfg.MaxOpenConns > 0 {
		sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	}
	if cfg.MaxIdleConns > 0 {
		sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	}
	if cfg.ConnMaxLifetime > 0 {
		sqlDB.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	}
	return db, nil
}

// autoMigrate runs only on the service DB. Do not include User/clas_users — that table lives in another DB and must not be created or altered by this service.
func autoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&entity.EventSource{},
		&entity.MetadataRule{},
		&entity.MetadataRuleAction{},
		&entity.ProcessedEvent{},
		&entity.FailedEvent{},
	)
}

// Ping checks connectivity.
func Ping(ctx context.Context, db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}
