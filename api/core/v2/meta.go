package v2

const (
	// ManagedByLabel is used to identify which client was used to create/update a
	// resource
	ManagedByLabel = "sensu.io/managed_by"
)

type Comparison int

const (
	MetaLess    Comparison = -1
	MetaEqual   Comparison = 0
	MetaGreater Comparison = 1
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

// NewObjectMetaP is the pointer-returning version of NewObjectMeta.
func NewObjectMetaP(name, namespace string) *ObjectMeta {
	meta := NewObjectMeta(name, namespace)
	return &meta
}

// Cmp compares this ObjectMeta with another ObjectMeta.
func (o *ObjectMeta) Cmp(other *ObjectMeta) Comparison {
	if o == nil {
		return MetaLess
	}
	if other == nil {
		return MetaGreater
	}
	if o.Namespace < other.Namespace {
		return MetaLess
	}
	if o.Namespace > other.Namespace {
		return MetaGreater
	}
	if o.Name < other.Name {
		return MetaLess
	}
	if o.Name > other.Name {
		return MetaGreater
	}
	return MetaEqual
}
