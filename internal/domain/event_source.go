package domain

import (
	"time"

	"github.com/google/uuid"
)

// EventSource represents a RabbitMQ topic/exchange to consume from.
type EventSource struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey"`
	TopicName string    `json:"topic_name" gorm:"column:topic_name;uniqueIndex"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at" gorm:"column:created_at"`
}

// TableName overrides table name.
func (EventSource) TableName() string {
	return "event_sources"
}
