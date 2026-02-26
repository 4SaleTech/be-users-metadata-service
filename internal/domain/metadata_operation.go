package domain

// MetadataOperation is a single key-level operation (set, increment, append, etc.) to apply to user metadata.
type MetadataOperation struct {
	Key   string
	Op    string
	Value interface{}
}
