package v3

// Redacter can return a redacted copy of the resource
type Redacter interface {
	// ProduceRedacted returns a redacted copy of the resource
	ProduceRedacted() Resource
}
