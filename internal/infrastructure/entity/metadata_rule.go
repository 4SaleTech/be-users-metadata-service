package entity

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MetadataRule represents metadata_rules table.
type MetadataRule struct {
	ID           uuid.UUID           `gorm:"type:char(36);primaryKey"`
	EventType    string              `gorm:"column:event_type;not null;index"`
	EventVersion string              `gorm:"column:event_version"`
	Enabled      bool                `gorm:"default:true"`
	Priority     int                 `gorm:"default:0"`
	Description  string              `gorm:"type:text"`
	CreatedAt    time.Time           `gorm:"autoCreateTime"`
	Actions      []MetadataRuleAction `gorm:"foreignKey:RuleID"`
}

// TableName overrides table name.
func (MetadataRule) TableName() string { return "metadata_rules" }

// BeforeCreate ensures UUID is set.
func (r *MetadataRule) BeforeCreate(tx *gorm.DB) error {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return nil
}
