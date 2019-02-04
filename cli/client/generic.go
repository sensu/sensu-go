package client

import (
	"encoding/json"
	"fmt"

	"github.com/sensu/sensu-go/types"
)

// Delete sends a DELETE request to the given path
func (client *RestClient) Delete(path string) error {
	res, err := client.R().Delete(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// Get sends a GET request for an object at the given path
func (client *RestClient) Get(path string, obj interface{}) error {
	res, err := client.R().SetResult(obj).Get(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// List sends a GET request for all objects at the given path
func (client *RestClient) List(path string, objs interface{}) error {
	res, err := client.R().Get(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return json.Unmarshal(res.Body(), objs)
}

// Post sends a POST request with obj as the payload to the given path
func (client *RestClient) Post(path string, obj interface{}) error {
	res, err := client.R().SetBody(obj).Post(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// PutResource ...
func (client *RestClient) PutResource(r types.Wrapper) error {
	path := r.Value.URIPath()

	// Determine if we should send the wrapped resource or only the resource
	// itself
	var bytes []byte
	var err error
	if r.APIVersion == "core/v2" {
		bytes, err = json.Marshal(r.Value)
	} else {
		bytes, err = json.Marshal(r)
	}
	if err != nil {
		return err
	}

	res, err := client.R().SetBody(bytes).Put(path)
	if err != nil {
		return fmt.Errorf("PUT %q: %s", path, err)
	}
	if res.StatusCode() >= 400 {
		return fmt.Errorf("PUT %q: %s", path, res.String())
	}
	return nil
}
