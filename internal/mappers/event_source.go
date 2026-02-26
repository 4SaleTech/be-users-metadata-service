package mappers

import (
	"github.com/be-users-metadata-service/internal/domain"
	"github.com/be-users-metadata-service/internal/infrastructure/entity"
)

// EventSourceToDomain maps entity to domain.
func EventSourceToDomain(e entity.EventSource) domain.EventSource {
	return domain.EventSource{
		ID:        e.ID,
		TopicName: e.TopicName,
		Enabled:   e.Enabled,
		CreatedAt: e.CreatedAt,
	}
}

// EventSourceFromDomain maps domain to entity.
func EventSourceFromDomain(d domain.EventSource) entity.EventSource {
	return entity.EventSource{
		ID:        d.ID,
		TopicName: d.TopicName,
		Enabled:   d.Enabled,
		CreatedAt: d.CreatedAt,
	}
}
