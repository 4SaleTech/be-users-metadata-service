package domain

// MetaDataMap is a convenience type for reading/writing metadata as map.
type MetaDataMap map[string]interface{}

// Merge applies another map on top (later keys override).
func (m MetaDataMap) Merge(other map[string]interface{}) {
	for k, v := range other {
		m[k] = v
	}
}
