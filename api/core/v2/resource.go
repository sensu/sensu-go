package v2

// Resource represents a Sensu resource.
type Resource interface {
	// GetObjectMeta returns the object metadata for the resource.
	GetObjectMeta() ObjectMeta

	// SetObjectMeta sets the object metadata for the resource.
	SetObjectMeta(ObjectMeta)

	// SetNamespace sets the namespace of the resource.
	SetNamespace(string)

	// StorePrefix gives the path prefix to this resource in the store
	StorePrefix() string

	// RBACName describes the name of the resource for RBAC purposes.
	RBACName() string

	// URIPath gives the path to the resource, e.g. /checks/checkname
	URIPath() string

	// Validate checks if the fields in the resource are valid.
	Validate() error
}
