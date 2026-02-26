package service

import (
	"context"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"

	"github.com/be-users-metadata-service/internal/application/util"
	"github.com/be-users-metadata-service/internal/domain"
)

// RuleEngine evaluates conditions, resolves values, and evaluates rules to produce metadata operations.
type RuleEngine struct{}

// NewRuleEngine returns a new RuleEngine.
func NewRuleEngine() *RuleEngine {
	return &RuleEngine{}
}

// EvaluateCondition returns true if condition is empty or evaluates to true.
func (e *RuleEngine) EvaluateCondition(ctx context.Context, condition string, event *domain.Event, metadata map[string]interface{}) (bool, error) {
	_ = ctx
	condition = strings.TrimSpace(condition)
	if condition == "" {
		return true, nil
	}
	ctxMap := buildContext(event, metadata)
	return evalSimpleCondition(condition, ctxMap), nil
}

// ResolveValue resolves value_template against event and metadata.
func (e *RuleEngine) ResolveValue(ctx context.Context, valueSource, valueTemplate string, event *domain.Event, metadata map[string]interface{}) (interface{}, error) {
	_ = ctx
	ctxMap := buildContext(event, metadata)
	switch valueSource {
	case domain.ValueSourceStatic:
		return parseValue(strings.TrimSpace(valueTemplate)), nil
	case domain.ValueSourceEvent:
		return getByPath(ctxMap, "event."+valueTemplate), nil
	case domain.ValueSourceMetadata:
		return getByPath(ctxMap, "metadata."+valueTemplate), nil
	default:
		return getByPath(ctxMap, "event."+valueTemplate), nil
	}
}

// EvaluateRules evaluates all rules and returns metadata operations (updates metaMap in place).
func (e *RuleEngine) EvaluateRules(ctx context.Context, event *domain.Event, rules []domain.MetadataRule, metaMap map[string]interface{}) ([]domain.MetadataOperation, error) {
	var operations []domain.MetadataOperation
	for _, rule := range rules {
		for _, action := range rule.Actions {
			ok, err := e.EvaluateCondition(ctx, action.ConditionExpression, event, metaMap)
			if err != nil || !ok {
				continue
			}
			val, err := e.ResolveValue(ctx, action.ValueSource, action.ValueTemplate, event, metaMap)
			if err != nil {
				continue
			}
			op := domain.MetadataOperation{Key: action.MetadataKey, Op: action.Operation, Value: val}
			operations = append(operations, op)
			applyOpToMap(metaMap, op)
		}
	}
	return operations, nil
}

func buildContext(event *domain.Event, metadata map[string]interface{}) map[string]interface{} {
	ctx := make(map[string]interface{})
	if event != nil {
		ctx["event"] = map[string]interface{}{
			"id":   event.ID,
			"type": event.Type,
			"data": rawMessageToMap(event.Data),
		}
	}
	if metadata != nil {
		ctx["metadata"] = metadata
	}
	return ctx
}

func rawMessageToMap(data json.RawMessage) map[string]interface{} {
	if len(data) == 0 {
		return nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		return nil
	}
	return m
}

var condRegex = regexp.MustCompile(`^(\S+)\s*(==|!=|>=|<=|>|<)\s*(.+)$`)

func evalSimpleCondition(condition string, ctx map[string]interface{}) bool {
	parts := condRegex.FindStringSubmatch(condition)
	if len(parts) != 4 {
		return false
	}
	path, op, rhsStr := strings.TrimSpace(parts[1]), parts[2], strings.TrimSpace(parts[3])
	lhs := getByPath(ctx, path)
	rhs := parseValue(rhsStr)
	return compare(lhs, op, rhs)
}

func getByPath(m map[string]interface{}, path string) interface{} {
	keys := strings.Split(path, ".")
	for _, k := range keys {
		if m == nil {
			return nil
		}
		v, ok := m[k]
		if !ok {
			return nil
		}
		if sub, ok := v.(map[string]interface{}); ok {
			m = sub
		} else {
			return v
		}
	}
	return m
}

func parseValue(s string) interface{} {
	s = strings.Trim(s, `"'`)
	if n, err := strconv.ParseInt(s, 10, 64); err == nil {
		return n
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}
	return s
}

func compare(lhs interface{}, op string, rhs interface{}) bool {
	switch op {
	case "==":
		return eq(lhs, rhs)
	case "!=":
		return !eq(lhs, rhs)
	case ">":
		return toFloat(lhs) > toFloat(rhs)
	case ">=":
		return toFloat(lhs) >= toFloat(rhs)
	case "<":
		return toFloat(lhs) < toFloat(rhs)
	case "<=":
		return toFloat(lhs) <= toFloat(rhs)
	default:
		return false
	}
}

func eq(a, b interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return toFloat(a) == toFloat(b) || toString(a) == toString(b)
}

func toFloat(v interface{}) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case string:
		f, _ := strconv.ParseFloat(x, 64)
		return f
	default:
		return 0
	}
}

func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return strconv.FormatFloat(toFloat(v), 'f', -1, 64)
}

func applyOpToMap(meta map[string]interface{}, op domain.MetadataOperation) {
	switch op.Op {
	case domain.OpSet:
		meta[op.Key] = op.Value
	case domain.OpIncrement:
		meta[op.Key] = util.ToFloat(meta[op.Key]) + util.ToFloat(op.Value)
	case domain.OpAppend:
		slice, _ := meta[op.Key].([]interface{})
		if slice == nil {
			slice = []interface{}{}
		}
		meta[op.Key] = append(slice, op.Value)
	case domain.OpRemove:
		delete(meta, op.Key)
	case domain.OpMerge:
		existing, _ := meta[op.Key].(map[string]interface{})
		if existing == nil {
			existing = make(map[string]interface{})
		}
		if newMap, ok := op.Value.(map[string]interface{}); ok {
			for k, v := range newMap {
				existing[k] = v
			}
		}
		meta[op.Key] = existing
	case domain.OpMax:
		a, b := util.ToFloat(meta[op.Key]), util.ToFloat(op.Value)
		if b > a {
			meta[op.Key] = b
		}
	case domain.OpMin:
		a, b := util.ToFloat(meta[op.Key]), util.ToFloat(op.Value)
		if a == 0 || b < a {
			meta[op.Key] = b
		}
	default:
		meta[op.Key] = op.Value
	}
}
