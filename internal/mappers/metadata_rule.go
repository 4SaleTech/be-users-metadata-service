package mappers

import (
	"sort"

	"github.com/be-users-metadata-service/internal/domain"
	"github.com/be-users-metadata-service/internal/infrastructure/entity"
)

// MetadataRuleToDomain maps entity to domain.
func MetadataRuleToDomain(r entity.MetadataRule) domain.MetadataRule {
	rule := domain.MetadataRule{
		ID:           r.ID,
		EventType:    r.EventType,
		EventVersion: r.EventVersion,
		Enabled:      r.Enabled,
		Priority:     r.Priority,
		Description:  r.Description,
		CreatedAt:    r.CreatedAt,
		Actions:      make([]domain.MetadataRuleAction, len(r.Actions)),
	}
	for idx := range r.Actions {
		rule.Actions[idx] = MetadataRuleActionToDomain(r.Actions[idx])
	}
	sort.Slice(rule.Actions, func(i, j int) bool {
		if rule.Actions[i].ExecutionOrder != rule.Actions[j].ExecutionOrder {
			return rule.Actions[i].ExecutionOrder < rule.Actions[j].ExecutionOrder
		}
		return rule.Actions[i].ID.String() < rule.Actions[j].ID.String()
	})
	return rule
}

// MetadataRuleFromDomain maps domain to entity.
func MetadataRuleFromDomain(d domain.MetadataRule) entity.MetadataRule {
	r := entity.MetadataRule{
		ID:           d.ID,
		EventType:    d.EventType,
		EventVersion: d.EventVersion,
		Enabled:      d.Enabled,
		Priority:     d.Priority,
		Description:  d.Description,
		CreatedAt:    d.CreatedAt,
		Actions:      make([]entity.MetadataRuleAction, len(d.Actions)),
	}
	for idx := range d.Actions {
		r.Actions[idx] = MetadataRuleActionFromDomain(d.Actions[idx])
	}
	return r
}
