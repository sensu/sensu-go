package client

import (
	"fmt"
)

// FetchClusterID fetches the sensu cluster id
func (client *RestClient) FetchClusterID() (string, error) {
	res, err := client.R().Get("/id")
	if err != nil {
		return "", fmt.Errorf("GET %q: %s", "/id", err)
	}

	if res.StatusCode() >= 400 {
		return "", UnmarshalError(res)
	}

	return string(res.Body()), err
}
