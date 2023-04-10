package client

import (
	"encoding/json"
	"net/http"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/types"
)

// PostAPIKey sends a POST request with obj as the payload to the given path
// and returns the location header of the key.
func (client *RestClient) PostAPIKey(path string, obj interface{}) (corev2.APIKeyResponse, error) {
	var response corev2.APIKeyResponse

	w := types.WrapResource(obj)
	res, err := client.R().SetBody(w).Post(path)
	if err != nil {
		return response, err
	}

	if res.StatusCode() >= 400 {
		return response, UnmarshalError(res)
	}

	if err := json.Unmarshal(res.Body(), &response); err != nil {
		return response, err
	}

	return response, nil
}

// CreateAPIKey creates a new api-key.
func (client *RestClient) CreateAPIKey(name, username string, hash []byte) (corev2.APIKeyResponse, error) {

	apikey := &corev2.APIKey{
		ObjectMeta: corev2.ObjectMeta{
			Name:        name,
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		Username: username,
		Hash:     hash,
	}

	return client.PostAPIKey(apikey.URIPath(), apikey)
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

	var wrapped []types.Wrapper
	err = json.Unmarshal(resp.Body(), &wrapped)
	if err != nil {
		return nil, err
	}
	result := make([]corev2.APIKey, len(wrapped))
	for i := range wrapped {
		result[i] = *wrapped[i].Value.(*corev2.APIKey)
	}
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
