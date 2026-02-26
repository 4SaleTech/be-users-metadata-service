package mappers

import (
	"github.com/be-users-metadata-service/internal/domain"
	"github.com/be-users-metadata-service/internal/infrastructure/entity"
	"gorm.io/datatypes"
)

// ProcessedEventToDomain maps entity to domain.
func ProcessedEventToDomain(e entity.ProcessedEvent) domain.ProcessedEvent {
	return domain.ProcessedEvent{
		EventID:     e.EventID,
		EventJSON:   []byte(e.EventJSON),
		ProcessedAt: e.ProcessedAt,
	}
}

// ProcessedEventFromDomain maps domain to entity.
func ProcessedEventFromDomain(d domain.ProcessedEvent) entity.ProcessedEvent {
	return entity.ProcessedEvent{
		EventID:     d.EventID,
		EventJSON:   datatypes.JSON(d.EventJSON),
		ProcessedAt: d.ProcessedAt,
	}
}
