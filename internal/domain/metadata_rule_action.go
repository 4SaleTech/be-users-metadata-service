package domain

import "github.com/google/uuid"

// MetadataRuleAction defines a single metadata operation.
type MetadataRuleAction struct {
	ID                  uuid.UUID `json:"id"`
	RuleID              uuid.UUID `json:"rule_id"`
	Operation           string    `json:"operation"`             // set, increment, append, remove, merge, max, min
	MetadataKey         string    `json:"metadata_key"`
	ValueSource         string    `json:"value_source"`          // static, event, metadata
	ValueTemplate       string    `json:"value_template"`       // e.g. data.amount
	ConditionExpression string    `json:"condition_expression"`  // optional
	ExecutionOrder      int       `json:"execution_order"`
}

// Operation constants for metadata actions.
const (
	OpSet       = "set"
	OpIncrement = "increment"
	OpAppend    = "append"
	OpRemove    = "remove"
	OpMerge     = "merge"
	OpMax       = "max"
	OpMin       = "min"
)

// ValueSource constants.
const (
	ValueSourceStatic   = "static"
	ValueSourceEvent    = "event"
	ValueSourceMetadata = "metadata"
)
