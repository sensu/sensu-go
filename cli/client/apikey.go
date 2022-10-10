package client

import (
	"encoding/json"
	"net/http"
	"strings"

	corev2 "github.com/sensu/core/v2"
)

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

// CreateAPIKey creates a new api-key.
func (client *RestClient) CreateAPIKey(username string) (string, error) {

	apikey := &corev2.APIKey{
		Username: username,
	}

	location, err := client.PostAPIKey(apikey.URIPath(), apikey)
	if err != nil {
		return "", err
	}

	location_arr := strings.Split(location, "/")
	result := location_arr[len(location_arr)-1]

	return result, nil
}

// DeleteAPIKey deletes an api-key.
func (client *RestClient) DeleteAPIKey(name string) error {
	apikey := &corev2.APIKey{
		ObjectMeta: corev2.ObjectMeta{
			Name: name,
		},
	}
	return client.Delete(apikey.URIPath())
}

// ListAPIKeys fetches all api-key from configured Sensu instance
func (client *RestClient) ListAPIKeys(options *ListOptions, header *http.Header) ([]corev2.APIKey, error) {
	apikey := &corev2.APIKey{}
	request := client.R()
	ApplyListOptions(request, options)
	resp, err := request.Get(apikey.URIPath())
	if err != nil {
		return nil, err
	}
	*header = resp.Header()
	if resp.StatusCode() >= 400 {
		return nil, UnmarshalError(resp)
	}

	var result []corev2.APIKey
	err = json.Unmarshal(resp.Body(), &result)
	return result, err
}

// FetchAPIKey fetches an api-key from configured Sensu instance
func (client *RestClient) FetchAPIKey(name string) (*corev2.APIKey, error) {
	apikey := &corev2.APIKey{
		ObjectMeta: corev2.ObjectMeta{
			Name: name,
		},
	}

	resp, err := client.R().Get(apikey.URIPath())
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return nil, UnmarshalError(resp)
	}
	var result corev2.APIKey
	return &result, json.Unmarshal(resp.Body(), &result)
}

// UpdateAPIKey update an api-key from configured Sensu instance
func (client *RestClient) UpdateAPIKey(name, username string) error {
	apikey := &corev2.APIKey{
		ObjectMeta: corev2.ObjectMeta{
			Name: name,
		},
	}

	obj := &corev2.APIKey{
		Username: username,
	}
	request := client.R()
	request.Header.Add("Content-Type", "application/merge-patch+json")
	_, err := request.SetBody(obj).Patch(apikey.URIPath())
	return err
}
