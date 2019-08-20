package client

import (
	"encoding/json"
	"fmt"

	"github.com/sensu/sensu-go/types"
)

// ExtPath is the api path for extensions.
var ExtPath = createNSBasePath(coreAPIGroup, coreAPIVersion, "extensions")

// DeregisterExtension deregisters an extension resource from the backend
func (client *RestClient) DeregisterExtension(name, namespace string) error {
	path := ExtPath(namespace, name)
	res, err := client.R().Delete(path)
	if err != nil {
		return fmt.Errorf("DELETE %q: %s", path, err)
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// RegisterExtension updates an extension resource from the backend
func (client *RestClient) RegisterExtension(extension *types.Extension) error {
	bytes, err := json.Marshal(extension)
	if err != nil {
		return err
	}

	path := ExtPath(extension.Namespace, extension.Name)
	res, err := client.R().SetBody(bytes).Put(path)
	if err != nil {
		return fmt.Errorf("PUT %q: %s", path, err)
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}
