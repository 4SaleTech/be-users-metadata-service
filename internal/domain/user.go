package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// User is the aggregate for user metadata updates (domain model; persistence in infrastructure).
type User struct {
	ID        uuid.UUID       `json:"id"`
	MetaData  json.RawMessage `json:"meta_data"`
	UpdatedAt time.Time       `json:"updated_at"`
}
