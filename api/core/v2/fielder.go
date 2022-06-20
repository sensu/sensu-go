package v2

// Fielder includes a set of fields that represent a resource.
type Fielder interface {
	// Fields returns a set of fields that represent the resource.
	Fields() map[string]string
}
