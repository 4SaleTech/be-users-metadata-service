package mappers

import (
	"github.com/be-users-metadata-service/internal/domain"
	"github.com/be-users-metadata-service/internal/infrastructure/entity"
	"gorm.io/datatypes"
)

// FailedEventToDomain maps entity to domain.
func FailedEventToDomain(e entity.FailedEvent) domain.FailedEvent {
	return domain.FailedEvent{
		ID:           e.ID,
		EventType:    e.EventType,
		Payload:      []byte(e.Payload),
		ErrorMessage: e.ErrorMessage,
		CreatedAt:    e.CreatedAt,
		ProcessedAt:  e.ProcessedAt,
	}
}

// FailedEventFromDomain maps domain to entity.
func FailedEventFromDomain(d domain.FailedEvent) entity.FailedEvent {
	return entity.FailedEvent{
		ID:           d.ID,
		EventType:    d.EventType,
		Payload:      datatypes.JSON(d.Payload),
		ErrorMessage: d.ErrorMessage,
		CreatedAt:    d.CreatedAt,
		ProcessedAt:  d.ProcessedAt,
	}
}
