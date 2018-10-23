package meta

import (
	"fmt"
	"net/url"
)

const (
	// uniqueRestKeyTmpl is the template for a uniquely identified Sensu object
	// /apis/{{ .GetGroup }}/{{ .GetVersion }}/namespaces/{{ .GetNamespace }/{{ .GetKind }}/{{ .GetName }}
	uniqueRestKeyTmpl = "/apis/%s/%s/namespaces/%s/%s/%s"

	// restKeyTmpl is the template for a given kind of Sensu object
	// /apis/{{ .GetGroup }}/{{ .GetVersion }}/namespaces/{{ .GetNamespace }/{{ .GetKind }}
	restKeyTmpl = "/apis/%s/%s/namespaces/%s/%s"

	// storageKeyTmpl is...
	// /sensu.io/{{ .GetGroup }}/{{ .GetVersion }}/{{ .GetNamespace }}/{{ .GetKind }}/{{ .GetName }}
	storageKeyTmpl = "/sensu.io/%s/%s/%s/%s/%s"

	// uniqueStorageKeyTmpl is...
	// /sensu.io/{{ .GetGroup }}/{{ .GetVersion }}/{{ .GetNamespace }}/{{ .GetKind }}
	uniqueStorageKeyTmpl = "/sensu.io/%s/%s/%s/%s"
)

// Keyable specifies the requirements for uniquely identifying a Sensu object.
type Keyable interface {
	// GetKind returns the kind of the object. (e.g. check, entity)
	GetKind() string

	// GetGroup returns the API group (e.g., rbac).
	GetGroup() string

	// GetVersion returns the API version (e.g., v1).
	GetVersion() string

	// GetNamespace returns the namespace the object lives under.
	GetNamespace() string

	// GetName returns the name of the object.
	GetName() string
}

// RESTKey is a data type which is a responsible for computing REST paths for
// Keyables.
type RESTKey struct {
	Keyable
}

// NewRESTKey creates a new RESTKey from TypeMeta and ObjectMeta data.
func NewRESTKey(tm TypeMeta, om ObjectMeta) RESTKey {
	return RESTKey{struct {
		TypeMeta
		*ObjectMeta
	}{tm, &om}}
}

func escapeSprintf(format string, params ...string) string {
	newParams := make([]interface{}, len(params))
	for i := range params {
		newParams[i] = url.PathEscape(params[i])
	}
	return fmt.Sprintf(format, newParams...)
}

func (r RESTKey) single() string {
	return escapeSprintf(uniqueRestKeyTmpl, r.GetGroup(), r.GetVersion(), r.GetNamespace(), r.GetKind(), r.GetName())
}

func (r RESTKey) multi() string {
	return escapeSprintf(restKeyTmpl, r.GetGroup(), r.GetVersion(), r.GetNamespace(), r.GetKind())
}

// GetPath returns the path to use for GET requests.
// Example: /apis/core/v1/namespaces/default/checks
func (r RESTKey) GetPath() string {
	return r.single()
}

// PutPath returns the path to use for PUT requests.
func (r RESTKey) PutPath() string {
	return r.single()
}

// PatchPath returns the path to use for PATCH requests.
func (r RESTKey) PatchPath() string {
	return r.single()
}

// DeletePath returns the path to use for DELETE requests.
func (r RESTKey) DeletePath() string {
	return r.single()
}

// ListPath returns the path to use for list requests. (GET without unique key)
func (r RESTKey) ListPath() string {
	return r.multi()
}

// PostPath returns the path to use for POST requests.
func (r RESTKey) PostPath() string {
	return r.multi()
}

// RESTKey is a data type which is a responsible for computing key/value
// storage key paths for Keyables.
type StorageKey struct {
	Keyable
}

// NewStorageKey creates a new StorageKey from TypeMeta and ObjectMeta data.
func NewStorageKey(tm TypeMeta, om ObjectMeta) StorageKey {
	return StorageKey{struct {
		TypeMeta
		*ObjectMeta
	}{tm, &om}}
}

// Path returns a unique key.
// Example: /sensu.io/core/v1/default/check/cpu
func (s StorageKey) Path() string {
	return fmt.Sprintf(storageKeyTmpl, s.GetGroup(), s.GetVersion(), s.GetNamespace(), s.GetKind(), s.GetName())
}

// PrefixPath returns a prefix path for the data type.
// Example: /sensu.io/core/v1/default/check
func (s StorageKey) PrefixPath() string {
	return fmt.Sprintf(uniqueStorageKeyTmpl, s.GetGroup(), s.GetVersion(), s.GetNamespace(), s.GetKind())
}
