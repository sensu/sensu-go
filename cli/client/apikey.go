package client

// PostAPIKey sends a POST request with obj as the payload to the given path
// and returns the location header of the key.
func (client *RestClient) PostAPIKey(path string, obj interface{}) (string, error) {
	res, err := client.R().SetBody(obj).Post(path)
	if err != nil {
		return "", err
	}

	if res.StatusCode() >= 400 {
		return "", UnmarshalError(res)
	}

	return res.Header().Get("Location"), nil
}
