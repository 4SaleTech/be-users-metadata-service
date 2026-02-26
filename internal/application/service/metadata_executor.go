package service

import (
	"context"
	"encoding/json"

	"github.com/be-users-metadata-service/internal/application/util"
	"github.com/be-users-metadata-service/internal/domain"
	"gorm.io/datatypes"
)

// MetadataExecutor applies metadata operations to produce the final JSON value.
type MetadataExecutor struct{}

// NewMetadataExecutor returns a new MetadataExecutor.
func NewMetadataExecutor() *MetadataExecutor {
	return &MetadataExecutor{}
}

// ComputeResult applies operations to currentMeta and returns new JSON.
func (m *MetadataExecutor) ComputeResult(ctx context.Context, currentMeta datatypes.JSON, operations []domain.MetadataOperation) (datatypes.JSON, error) {
	_ = ctx
	var meta map[string]interface{}
	if len(currentMeta) > 0 {
		if err := json.Unmarshal(currentMeta, &meta); err != nil {
			meta = make(map[string]interface{})
		}
	}
	if meta == nil {
		meta = make(map[string]interface{})
	}
	for _, op := range operations {
		applyOp(meta, op)
	}
	b, err := json.Marshal(meta)
	if err != nil {
		return nil, err
	}
	return datatypes.JSON(b), nil
}

func applyOp(meta map[string]interface{}, op domain.MetadataOperation) {
	switch op.Op {
	case domain.OpSet:
		meta[op.Key] = op.Value
	case domain.OpIncrement:
		v := util.ToFloat(op.Value) + util.ToFloat(meta[op.Key])
		meta[op.Key] = v
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
