package v2

// Resource represents a Sensu resource.
type Resource interface {
	// URIPath gives the path to the resource, e.g. /checks/checkname
	URIPath() string

	// Validate checks if the fields in the resource are valid.
	Validate() error

	// GetObjectMeta returns the object metadata for the resource.
	GetObjectMeta() ObjectMeta
}

// AbstractResource is a resource holder in cases we only care about the
// resource metadata but we need to pass that information as a Resource
type AbstractResource struct {
	ObjectMeta
}

// GetObjectMeta only exists here to fulfil the requirements of Resource
func (r *AbstractResource) GetObjectMeta() ObjectMeta {
	return r.ObjectMeta
}

// URIPath returns an empty string
func (r *AbstractResource) URIPath() string {
	return ""
}

// Validate returns no error
func (r *AbstractResource) Validate() error {
	return nil
}
