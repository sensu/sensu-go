package v2

import (
	"fmt"
	"net/url"
	"path"

	"github.com/google/uuid"
	stringsutil "github.com/sensu/sensu-go/api/core/v2/internal/stringutil"
)

const (
	// APIKeysResource is the name of this resource type.
	APIKeysResource = "apikeys"
)

// StorePrefix returns the path prefix to this resource in the store.
func (a *APIKey) StorePrefix() string {
	return APIKeysResource
}

// URIPath returns the path component of an api key URI.
func (a *APIKey) URIPath() string {
	return path.Join(URLPrefix, APIKeysResource, url.PathEscape(a.Name))
}

// Validate returns an error if the CheckName and Subscription fields are not
// provided.
func (a *APIKey) Validate() error {
	if a.Namespace != "" {
		return fmt.Errorf("api key cannot have a namespace")
	}

	if a.Username == "" {
		return fmt.Errorf("api key must have a username")
	}

	if _, err := uuid.Parse(a.Name); err != nil {
		return fmt.Errorf("api key name: %s", err)
	}

	return nil
}

// FixtureAPIKey returns a testing fixture for an APIKey struct.
func FixtureAPIKey(name string, username string) *APIKey {
	return &APIKey{
		Username:   username,
		ObjectMeta: NewObjectMeta(name, ""),
	}
}

// APIKeyFields returns a set of fields that represent that resource.
func APIKeyFields(r Resource) map[string]string {
	resource := r.(*APIKey)
	fields := map[string]string{
		"api_key.name":     resource.ObjectMeta.Name,
		"api_key.username": resource.Username,
	}
	stringsutil.MergeMapWithPrefix(fields, resource.ObjectMeta.Labels, "api_key.labels.")
	return fields
}

// SetNamespace sets the namespace of the resource.
func (a *APIKey) SetNamespace(namespace string) {}

// SetObjectMeta sets the meta of the resource.
func (a *APIKey) SetObjectMeta(meta ObjectMeta) {
	a.ObjectMeta = meta
}

// RBACName gets the rbac name of the resource.
func (*APIKey) RBACName() string {
	return APIKeysResource
}
