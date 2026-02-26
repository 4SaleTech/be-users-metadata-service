package mappers

import (
	"github.com/be-users-metadata-service/internal/domain"
	"github.com/be-users-metadata-service/internal/infrastructure/entity"
)

// MetadataRuleActionToDomain maps entity to domain.
func MetadataRuleActionToDomain(a entity.MetadataRuleAction) domain.MetadataRuleAction {
	return domain.MetadataRuleAction{
		ID:                  a.ID,
		RuleID:              a.RuleID,
		Operation:           a.Operation,
		MetadataKey:         a.MetadataKey,
		ValueSource:         a.ValueSource,
		ValueTemplate:       a.ValueTemplate,
		ConditionExpression: a.ConditionExpression,
		ExecutionOrder:      a.ExecutionOrder,
	}
}

// MetadataRuleActionFromDomain maps domain to entity.
func MetadataRuleActionFromDomain(d domain.MetadataRuleAction) entity.MetadataRuleAction {
	return entity.MetadataRuleAction{
		ID:                  d.ID,
		RuleID:              d.RuleID,
		Operation:           d.Operation,
		MetadataKey:         d.MetadataKey,
		ValueSource:         d.ValueSource,
		ValueTemplate:       d.ValueTemplate,
		ConditionExpression: d.ConditionExpression,
		ExecutionOrder:      d.ExecutionOrder,
	}
}
