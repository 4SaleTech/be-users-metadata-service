package entity

import (
	"encoding/json"

	"gorm.io/datatypes"
)

// MetaDataMap type for JSON operations.
type MetaDataMap map[string]interface{}

// ToJSON returns datatypes.JSON from map.
func (m MetaDataMap) ToJSON() (datatypes.JSON, error) {
	if m == nil {
		return []byte("{}"), nil
	}
	b, err := json.Marshal(m)
	return datatypes.JSON(b), err
}

// Scan implements sql.Scanner for reading JSON into map.
func (m *MetaDataMap) Scan(value interface{}) error {
	if value == nil {
		*m = nil
		return nil
	}
	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return nil
	}
	return json.Unmarshal(b, m)
}
