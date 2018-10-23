package storage

import (
	"fmt"
	"net/url"
)

// Most objects follow the path patterns here.
const (
	// uniqueRestKeyTmpl is the template for a uniquely identified Sensu
	// object that has a namespace.
	// /apis/{{ .GetGroup }}/{{ .GetVersion }}/namespaces/{{ .GetNamespace }/{{ .GetKind }}/{{ .GetName }}
	uniqueRestKeyTmpl = "/apis/%s/%s/namespaces/%s/%s/%s"

	// restKeyTmpl is the template for a list of Sensu objects that have a
	// namespace.
	// /apis/{{ .GetGroup }}/{{ .GetVersion }}/namespaces/{{ .GetNamespace }/{{ .GetKind }}
	restKeyTmpl = "/apis/%s/%s/namespaces/%s/%s"

	// storageKeyTmpl is a template for a prefix identifier for an object
	// that has a namespace.
	// /sensu.io/{{ .GetGroup }}/{{ .GetVersion }}/{{ .GetNamespace }}/{{ .GetKind }}/{{ .GetName }}
	uniqueStorageKeyTmpl = "/sensu.io/%s/%s/%s/%s/%s"

	// uniqueStorageKeyTmpl is a template for a unique identifier for an object
	// that has a namespace.
	// /sensu.io/{{ .GetGroup }}/{{ .GetVersion }}/{{ .GetNamespace }}/{{ .GetKind }}
	storageKeyTmpl = "/sensu.io/%s/%s/%s/%s"
)

// Some special objects, that don't have namespaces, follow the path patterns here.
const (
	// globalUniqueRestKeyTmpl is the template for a uniquely identified Sensu
	// object that does not have a namespace.
	// /apis/{{ .GetGroup }}/{{ .GetVersion }}/{{ .GetKind }}/{{ .GetName }}
	globalUniqueRestKeyTmpl = "/apis/%s/%s/%s/%s"

	// globalRestKeyTmpl is the template for a list of Sensu objects that
	// do not have a namespace.
	// /apis/{{ .GetGroup }}/{{ .GetVersion }}/{{ .GetKind }}
	globalRestKeyTmpl = "/apis/%s/%s/%s"

	// globalStorageKeyTmpl is a template for a prefix identifier for an object
	// that does not have a namespace.
	// /sensu.io/{{ .GetGroup }}/{{ .GetVersion }}/{{ .GetKind }}
	globalStorageKeyTmpl = "/sensu.io/%s/%s/%s"

	// globalUniqueStorageKeyTmpl is a template for a unique identifier for an
	// object that does not have a namespace.
	// /sensu.io/{{ .GetGroup }}/{{ .GetVersion }}/{{ .GetKind }}/{{ .GetName }}
	globalUniqueStorageKeyTmpl = "/sensu.io/%s/%s/%s/%s"
)

// noNamespace is a way to detect if a type does not belong in a namespace.
type noNamespace interface {
	NoNamespace()
}

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

func escapeSprintf(format string, params ...string) string {
	newParams := make([]interface{}, len(params))
	for i := range params {
		newParams[i] = url.PathEscape(params[i])
	}
	return fmt.Sprintf(format, newParams...)
}

func (r RESTKey) single() string {
	tmpl := globalUniqueRestKeyTmpl
	args := []string{r.GetGroup(), r.GetVersion()}

	if isNamespaced(r.Keyable) {
		args = append(args, r.GetNamespace())
		tmpl = uniqueRestKeyTmpl
	}

	args = append(args, r.GetKind(), r.GetName())

	return escapeSprintf(tmpl, args...)
}

func (r RESTKey) multi() string {
	tmpl := globalRestKeyTmpl
	args := []string{r.GetGroup(), r.GetVersion()}

	if isNamespaced(r.Keyable) {
		args = append(args, r.GetNamespace())
		tmpl = restKeyTmpl
	}
	args = append(args, r.GetKind())

	return escapeSprintf(tmpl, args...)
}

// GetPath returns the path to use for GET requests.
// Example: /apis/core/v1/namespaces/default/checks/cpu
func (r RESTKey) GetPath() string {
	return r.single()
}

// PutPath returns the path to use for PUT requests.
// Example: /apis/core/v1/namespaces/default/checks/cpu
func (r RESTKey) PutPath() string {
	return r.single()
}

// PatchPath returns the path to use for PATCH requests.
// Example: /apis/core/v1/namespaces/default/checks/cpu
func (r RESTKey) PatchPath() string {
	return r.single()
}

// DeletePath returns the path to use for DELETE requests.
// Example: /apis/core/v1/namespaces/default/checks/cpu
func (r RESTKey) DeletePath() string {
	return r.single()
}

// ListPath returns the path to use for list requests. (GET without unique key)
// Example: /apis/core/v1/namespaces/default/checks
func (r RESTKey) ListPath() string {
	return r.multi()
}

// PostPath returns the path to use for POST requests.
// Example: /apis/core/v1/namespaces/default/checks
func (r RESTKey) PostPath() string {
	return r.multi()
}

// RESTKey is a data type which is a responsible for computing key/value
// storage key paths for Keyables.
type StorageKey struct {
	Keyable
}

// discover if a keyable is meant to exist within a namespace
func isNamespaced(k Keyable) bool {
	_, ok := k.(noNamespace)
	return !ok
}

// Path returns a unique key.
// Example: /sensu.io/core/v1/default/check/cpu
func (s StorageKey) Path() string {
	args := []interface{}{s.GetGroup(), s.GetVersion()}
	var tmpl string

	if isNamespaced(s.Keyable) {
		args = append(args, s.GetNamespace())
		tmpl = uniqueStorageKeyTmpl
	} else {
		tmpl = globalUniqueStorageKeyTmpl
	}
	args = append(args, s.GetKind(), s.GetName())

	return fmt.Sprintf(tmpl, args...)
}

// PrefixPath returns a prefix path for the data type.
// Example: /sensu.io/core/v1/default/check
func (s StorageKey) PrefixPath() string {
	var tmpl string
	args := []interface{}{s.GetGroup(), s.GetVersion()}

	if isNamespaced(s.Keyable) {
		tmpl = storageKeyTmpl
		args = append(args, s.GetNamespace())
	} else {
		tmpl = globalStorageKeyTmpl
	}
	args = append(args, s.GetKind())

	return fmt.Sprintf(tmpl, args...)
}
