package domain

import (
	"time"

	"github.com/google/uuid"
)

// MetadataRule represents a rule that matches events and defines actions.
type MetadataRule struct {
	ID           uuid.UUID           `json:"id"`
	EventType    string              `json:"event_type"`
	EventVersion string              `json:"event_version"`
	Enabled      bool                `json:"enabled"`
	Priority     int                 `json:"priority"`
	Description  string              `json:"description"`
	CreatedAt    time.Time           `json:"created_at"`
	Actions      []MetadataRuleAction `json:"actions"`
}
