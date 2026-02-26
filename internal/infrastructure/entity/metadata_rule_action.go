package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MetadataRuleAction represents metadata_rule_actions table.
type MetadataRuleAction struct {
	ID                  uuid.UUID `gorm:"type:char(36);primaryKey"`
	RuleID              uuid.UUID `gorm:"type:char(36);not null;index"`
	Operation           string    `gorm:"not null"` // set, increment, append, remove, merge, max, min
	MetadataKey         string    `gorm:"column:metadata_key;not null"`
	ValueSource         string    `gorm:"column:value_source;not null"`
	ValueTemplate       string    `gorm:"column:value_template;type:text"`
	ConditionExpression string    `gorm:"column:condition_expression;type:text"`
	ExecutionOrder      int       `gorm:"column:execution_order;default:0"`
}

// TableName overrides table name.
func (MetadataRuleAction) TableName() string { return "metadata_rule_actions" }

// BeforeCreate ensures UUID is set.
func (a *MetadataRuleAction) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
