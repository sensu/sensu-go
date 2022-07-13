package handlers

import (
	"errors"
	"fmt"
	"net/url"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// Handlers represents the HTTP handlers for CRUD operations on resources
type Handlers struct {
	Resource corev3.Resource
	Store    storev2.Interface
}

func checkMeta(meta corev2.ObjectMeta, vars map[string]string, idVar string) error {
	namespace, err := url.PathUnescape(vars["namespace"])
	if err != nil {
		return err
	}

	if meta.Namespace != namespace && namespace != "" {
		return fmt.Errorf(
			"the namespace of the resource (%s) does not match the namespace of the URI (%s)",
			meta.Namespace,
			namespace,
		)
	}

	// The URL path name that holds the resource ID might differ, but fallback
	// to "id" if not provided
	if idVar == "" {
		idVar = "id"
	}

	id, err := url.PathUnescape(vars[idVar])
	if err != nil {
		return err
	}

	if meta.Name != id && id != "" {
		return fmt.Errorf(
			"the name of the resource (%s) does not match the name of the URI (%s)",
			meta.Name,
			id,
		)
	}

	return nil
}

// V3CheckMeta inspects the resource metadata and ensures it matches what was
// specified in the request URL. Unlike CheckMeta it operates on v3 resources.
func CheckV3Meta(resource interface{}, vars map[string]string, idVar string) error {
	v, ok := resource.(interface{ GetMetadata() *corev2.ObjectMeta })
	if !ok {
		// We are not dealing with a corev3.Resource interface
		return nil
	}
	meta := v.GetMetadata()
	if meta == nil {
		return errors.New("nil metadata")
	}
	return checkMeta(*meta, vars, idVar)
}

// CheckMeta inspects the resource metadata and ensures it matches what was
// specified in the request URL
func CheckMeta(resource interface{}, vars map[string]string, idVar string) error {
	v, ok := resource.(interface{ GetObjectMeta() corev2.ObjectMeta })
	if !ok {
		// We are not dealing with a corev2.Resource interface
		return nil
	}
	meta := v.GetObjectMeta()
	return checkMeta(meta, vars, idVar)
}

// Resource is used to set metadata values, e.g. in MetaPathValues()
type Resource interface {
	GetObjectMeta() corev2.ObjectMeta
	SetNamespace(string)
	SetName(string)
}

// MetaPathValues inspects the resource metadata and pulls values from
// the path variables when specific values are missing.
func MetaPathValues(resource Resource, muxVars map[string]string, nameVar string) error {
	meta := resource.GetObjectMeta()

	if meta.Namespace == "" {
		namespace, err := url.PathUnescape(muxVars["namespace"])
		if err != nil {
			return err
		}

		resource.SetNamespace(namespace)
	}

	if meta.Name == "" {
		if nameVar == "" {
			nameVar = "id"
		}

		name, err := url.PathUnescape(muxVars[nameVar])
		if err != nil {
			return err
		}

		resource.SetName(name)
	}

	return nil
}
