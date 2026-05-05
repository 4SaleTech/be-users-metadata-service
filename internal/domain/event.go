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
// If the JSON is a publish envelope with a nested "event" object (topic / routing_key / event),
// the inner object is unmarshalled so user_id and type are found.
func (e *Event) UnmarshalJSON(data []byte) error {
	payload := unwrapEnvelopeEventJSON(data)
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
	if err := json.Unmarshal(payload, &raw); err != nil {
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

// unwrapEnvelopeEventJSON returns inner "event" JSON when the payload looks like
// {"topic":"...","routing_key":"...","event":{...}}; otherwise returns data unchanged.
func unwrapEnvelopeEventJSON(data []byte) []byte {
	var envelope struct {
		Event json.RawMessage `json:"event"`
	}
	if err := json.Unmarshal(data, &envelope); err != nil || len(envelope.Event) == 0 || !json.Valid(envelope.Event) {
		return data
	}
	var probe map[string]interface{}
	if err := json.Unmarshal(envelope.Event, &probe); err != nil || len(probe) == 0 {
		return data
	}
	return envelope.Event
}

// EventID returns a stable idempotency key (e.g. from headers or body).
func (e *Event) EventID() string {
	if e.ID != "" {
		return e.ID
	}
	return uuid.New().String()
}
