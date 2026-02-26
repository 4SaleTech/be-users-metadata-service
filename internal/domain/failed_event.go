package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// FailedEvent stores events that could not be processed.
type FailedEvent struct {
	ID           uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey"`
	EventType    string          `json:"event_type" gorm:"column:event_type"`
	Payload      json.RawMessage `json:"payload"`
	ErrorMessage string          `json:"error_message" gorm:"column:error_message"`
	CreatedAt    time.Time       `json:"created_at" gorm:"column:created_at"`
	ProcessedAt  *time.Time      `json:"processed_at,omitempty" gorm:"column:processed_at"`
}

// TableName overrides table name.
func (FailedEvent) TableName() string {
	return "failed_events"
}
