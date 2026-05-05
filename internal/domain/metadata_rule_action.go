package domain

import "github.com/google/uuid"

// MetadataRuleAction defines a single metadata operation.
type MetadataRuleAction struct {
	ID                  uuid.UUID `json:"id"`
	RuleID              uuid.UUID `json:"rule_id"`
	Operation           string    `json:"operation"`            // set, increment, append, remove, merge, max, min
	MetadataKey         string    `json:"metadata_key"`         // literal or `${event.data.field}` / `${metadata.x}` template
	ValueSource         string    `json:"value_source"`         // static, event, metadata, formula
	ValueTemplate       string    `json:"value_template"`       // path, literal, or expression when value_source is formula
	ConditionExpression string    `json:"condition_expression"` // optional; paths like event.data.kind == "x"
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
	// ValueSourceFormula: value_template is a numeric expression (+-*/ and parentheses) using
	// dotted paths such as metadata.count or event.data.amount (same path roots as conditions).
	ValueSourceFormula = "formula"
)
