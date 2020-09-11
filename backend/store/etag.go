package store

// ETagCondition represents a conditional request
type ETagCondition struct {
	IfMatch     string
	IfNoneMatch string
}
