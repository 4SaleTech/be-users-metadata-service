package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// FailedEvent represents failed_events table.
type FailedEvent struct {
	ID           uuid.UUID      `gorm:"type:char(36);primaryKey"`
	EventType    string         `gorm:"column:event_type;not null"`
	Payload      datatypes.JSON `gorm:"type:json;not null"`
	ErrorMessage string         `gorm:"column:error_message;type:text"`
	CreatedAt    time.Time      `gorm:"column:created_at;autoCreateTime"`
	ProcessedAt  *time.Time     `gorm:"column:processed_at"`
}

// TableName overrides table name.
func (FailedEvent) TableName() string { return "failed_events" }

// BeforeCreate ensures UUID is set.
func (f *FailedEvent) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}
