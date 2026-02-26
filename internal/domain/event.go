package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Event represents an incoming message from RabbitMQ.
type Event struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Version   string          `json:"version,omitempty"`
	Source    string          `json:"source,omitempty"`
	Timestamp time.Time       `json:"timestamp,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
	UserID    *uuid.UUID      `json:"user_id,omitempty"`
}

// EventID returns a stable idempotency key (e.g. from headers or body).
func (e *Event) EventID() string {
	if e.ID != "" {
		return e.ID
	}
	return uuid.New().String()
}
