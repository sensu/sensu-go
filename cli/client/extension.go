package client

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/sensu/sensu-go/types"
)

// ListExtensions retrieves a list of extension resources from the backend
func (client *RestClient) ListExtensions(namespace string) ([]types.Extension, error) {
	var extensions []types.Extension

	res, err := client.R().Get("/extensions?namespace=" + namespace)
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
	extensionPath := fmt.Sprintf("/extensions/%s?namespace=%s", url.PathEscape(name), url.PathEscape(namespace))
	res, err := client.R().Delete(extensionPath)
	if err != nil {
		return fmt.Errorf("DELETE %q: %s", extensionPath, err)
	}

	if res.StatusCode() >= 400 {
		return fmt.Errorf("DELETE %q: %s", extensionPath, res.String())
	}

	return nil
}

// RegisterExtension updates an extension resource from the backend
func (client *RestClient) RegisterExtension(extension *types.Extension) error {
	bytes, err := json.Marshal(extension)
	if err != nil {
		return err
	}

	extensionPath := fmt.Sprintf("/extensions/%s", url.PathEscape(extension.Name))
	res, err := client.R().SetBody(bytes).Put(extensionPath)
	if err != nil {
		return fmt.Errorf("PUT %q: %s", extensionPath, err)
	}

	if res.StatusCode() >= 400 {
		return fmt.Errorf("PUT %q: %s", extensionPath, res.String())
	}

	return nil
}
