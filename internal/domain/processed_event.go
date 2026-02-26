package domain

import (
	"encoding/json"
	"time"
)

// ProcessedEvent stores event idempotency.
type ProcessedEvent struct {
	EventID     string          `json:"event_id" gorm:"primaryKey;column:event_id"`
	EventJSON   json.RawMessage `json:"event_json"`
	ProcessedAt time.Time       `json:"processed_at" gorm:"column:processed_at"`
}

// TableName overrides table name.
func (ProcessedEvent) TableName() string {
	return "processed_events"
}
