package entity

import (
	"time"

	"gorm.io/datatypes"
)

// ProcessedEvent represents processed_events table (idempotency).
type ProcessedEvent struct {
	EventID     string         `gorm:"column:event_id;primaryKey"`
	EventJSON   datatypes.JSON `gorm:"column:event_json;type:json;not null"`
	ProcessedAt time.Time      `gorm:"column:processed_at;not null"`
}

// TableName overrides table name.
func (ProcessedEvent) TableName() string { return "processed_events" }
