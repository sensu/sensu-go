package v2

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
