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

// NewDB creates a GORM DB for the primary (service) database and runs AutoMigrate for service tables only.
// Does not migrate User/clas_users; that table lives in the users DB (see NewUsersDB).
func NewDB(cfg Config) (*gorm.DB, error) {
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	configureConnPool(db, cfg)
	if err := autoMigratePrimary(db); err != nil {
		return nil, err
	}
	return db, nil
}

// NewUsersDB creates a GORM DB for the users database (clas_users table). No AutoMigrate.
// Use when DSN is non-empty (two-DB setup). Caller can use primary DB when DSN is empty (single-DB).
func NewUsersDB(cfg struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}) (*gorm.DB, error) {
	if cfg.DSN == "" {
		return nil, nil
	}
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

func configureConnPool(db *gorm.DB, cfg Config) {
	sqlDB, _ := db.DB()
	if sqlDB == nil {
		return
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
}

func autoMigratePrimary(db *gorm.DB) error {
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
