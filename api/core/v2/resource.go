package v2

// Resource represents a Sensu resource.
type Resource interface {
	// URIPath gives the path to the resource, e.g. /checks/checkname
	URIPath() string

	// Validate checks if the fields in the resource are valid.
	Validate() error

	// GetObjectMeta returns the object metadata for the resource.
	GetObjectMeta() ObjectMeta

	// SetNamespace sets the namespace of the resource.
	SetNamespace(string)
}
