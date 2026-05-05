package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Event represents an incoming message from RabbitMQ.
// user_id can be any string (e.g. numeric ID or UUID); event_type is accepted as alias for type.
type Event struct {
	ID        string          `json:"id"`
	Type      string          `json:"type"`
	Version   string          `json:"version,omitempty"`
	Source    string          `json:"source,omitempty"`
	Timestamp time.Time       `json:"timestamp,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
	UserID    string          `json:"user_id,omitempty"`
}

// UnmarshalJSON supports "event_type" as alias for "type" so payloads with event_type work.
func (e *Event) UnmarshalJSON(data []byte) error {
	var raw struct {
		ID        string          `json:"id"`
		Type      string          `json:"type"`
		EventType string          `json:"event_type"`
		Version   string          `json:"version,omitempty"`
		Source    string          `json:"source,omitempty"`
		Timestamp time.Time       `json:"timestamp,omitempty"`
		Data      json.RawMessage `json:"data,omitempty"`
		UserID    string          `json:"user_id,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	e.ID = raw.ID
	e.Type = raw.Type
	if e.Type == "" {
		e.Type = raw.EventType
	}
	e.Version = raw.Version
	e.Source = raw.Source
	e.Timestamp = raw.Timestamp
	e.Data = raw.Data
	e.UserID = raw.UserID
	return nil
}

// EventID returns a stable idempotency key (e.g. from headers or body).
func (e *Event) EventID() string {
	if e.ID != "" {
		return e.ID
	}
	return uuid.New().String()
}
