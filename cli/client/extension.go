package client

import (
	"encoding/json"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/types"
)

var extPath = createNSBasePath(coreAPIGroup, coreAPIVersion, "extensions")

// ListExtensions retrieves a list of extension resources from the backend
func (client *RestClient) ListExtensions(namespace string, options ListOptions) ([]corev2.Extension, error) {
	var extensions []corev2.Extension

	path := extPath(namespace)
	request := client.R()

	ApplyListOptions(request, options)

	res, err := request.Get(path)
	if err != nil {
		return extensions, err
	}

	if res.StatusCode() >= 400 {
		return extensions, fmt.Errorf("%v", res.String())
	}

	err = json.Unmarshal(res.Body(), &extensions)
	return extensions, err
}

// DeregisterExtension deregisters an extension resource from the backend
func (client *RestClient) DeregisterExtension(name, namespace string) error {
	path := extPath(namespace, name)
	res, err := client.R().Delete(path)
	if err != nil {
		return fmt.Errorf("DELETE %q: %s", path, err)
	}

	if res.StatusCode() >= 400 {
		return fmt.Errorf("DELETE %q: %s", path, res.String())
	}

	return nil
}

// RegisterExtension updates an extension resource from the backend
func (client *RestClient) RegisterExtension(extension *types.Extension) error {
	bytes, err := json.Marshal(extension)
	if err != nil {
		return err
	}

	path := extPath(extension.Namespace, extension.Name)
	res, err := client.R().SetBody(bytes).Put(path)
	if err != nil {
		return fmt.Errorf("PUT %q: %s", path, err)
	}

	if res.StatusCode() >= 400 {
		return fmt.Errorf("PUT %q: %s", path, res.String())
	}

	return nil
}
