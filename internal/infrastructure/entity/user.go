package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// User represents users table with meta_data JSON column.
type User struct {
	ID        uuid.UUID      `gorm:"type:char(36);primaryKey"`
	MetaData  datatypes.JSON `gorm:"column:meta_data;type:json"`
	UpdatedAt time.Time      `gorm:"column:updated_at;autoUpdateTime"`
}

// TableName overrides table name.
func (User) TableName() string { return "users" }
