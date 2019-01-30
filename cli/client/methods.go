package client

import "encoding/json"

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
