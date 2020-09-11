package store

// ETagCondition represents a conditional request
type ETagCondition struct {
	IfMatch     []byte
	IfNoneMatch []byte
}
