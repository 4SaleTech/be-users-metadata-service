package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EventSource represents event_sources table.
type EventSource struct {
	ID        uuid.UUID `gorm:"type:char(36);primaryKey"`
	TopicName string    `gorm:"column:topic_name;uniqueIndex;not null;size:512"`
	Enabled   bool      `gorm:"default:true"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// TableName overrides table name.
func (EventSource) TableName() string { return "event_sources" }

// BeforeCreate ensures UUID is set.
func (e *EventSource) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}
