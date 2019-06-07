package handlers

import (
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

// Handlers represents the HTTP handlers for CRUD operations on resources
type Handlers struct {
	Resource corev2.Resource
	Store    store.ResourceStore
}

// CheckMeta inspects the resource metadata and ensures it matches what was
// specified in the request URL
func CheckMeta(resource interface{}, vars map[string]string) error {
	v, ok := resource.(interface{ GetObjectMeta() corev2.ObjectMeta })
	if !ok {
		// We are not dealing with a corev2.Resource interface
		return nil
	}
	meta := v.GetObjectMeta()

	if meta.Namespace != vars["namespace"] && vars["namespace"] != "" {
		return fmt.Errorf(
			"the namespace of the resource (%s) does not match the namespace on the request (%s)",
			meta.Namespace,
			vars["namespace"],
		)
	}

	if meta.Name != vars["id"] && vars["id"] != "" {
		return fmt.Errorf(
			"the name of the resource (%s) does not match the name on the request (%s)",
			meta.Name,
			vars["id"],
		)
	}

	return nil
}
