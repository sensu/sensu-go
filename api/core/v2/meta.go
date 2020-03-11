package v2

const (
	// ManagedByLabel is used to identify which client was used to create/update a
	// resource
	ManagedByLabel = "sensu.io/managed_by"
)

// NewObjectMeta makes a new ObjectMeta, with Labels and Annotations assigned
// empty maps.
func NewObjectMeta(name, namespace string) ObjectMeta {
	return ObjectMeta{
		Name:        name,
		Namespace:   namespace,
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
	}
}
