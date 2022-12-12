package client

import (
	"encoding/json"
	"net/http"

	corev2 "github.com/sensu/core/v2"
)

var silencedPath = createNSBasePath(coreAPIGroup, coreAPIVersion, "silenced")

// CreateSilenced creates a new silenced  entry.
func (client *RestClient) CreateSilenced(silenced *corev2.Silenced) error {
	b, err := json.Marshal(silenced)
	if err != nil {
		return err
	}

	path := silencedPath(silenced.Namespace)
	res, err := client.R().SetBody(b).Post(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// DeleteSilenced deletes a silenced entry.
func (client *RestClient) DeleteSilenced(namespace, name string) error {
	return client.Delete(silencedPath(namespace, name))
}

// ListSilenceds fetches all silenced entries from configured Sensu instance
func (client *RestClient) ListSilenceds(namespace, sub, check string, options *ListOptions, header *http.Header) ([]corev2.Silenced, error) {
	if sub != "" && check != "" {
		name, err := corev2.SilencedName(sub, check)
		if err != nil {
			return nil, err
		}
		silenced, err := client.FetchSilenced(name)
		if err != nil {
			return nil, err
		}
		return []corev2.Silenced{*silenced}, nil
	}
	path := silencedPath(namespace)
	request := client.R()

	ApplyListOptions(request, options)

	if sub != "" {
		path = silencedPath(namespace, "subscriptions", sub)
	} else if check != "" {
		path = silencedPath(namespace, "checks", check)
	}
	resp, err := request.Get(path)
	if err != nil {
		return nil, err
	}
	*header = resp.Header()
	if resp.StatusCode() >= 400 {
		return nil, UnmarshalError(resp)
	}

	var result []corev2.Silenced
	err = json.Unmarshal(resp.Body(), &result)
	return result, err
}

// FetchSilenced fetches a silenced entry from configured Sensu instance
func (client *RestClient) FetchSilenced(name string) (*corev2.Silenced, error) {
	path := silencedPath(client.config.Namespace(), name)
	resp, err := client.R().Get(path)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return nil, UnmarshalError(resp)
	}
	var result corev2.Silenced
	return &result, json.Unmarshal(resp.Body(), &result)
}

// UpdateSilenced updates a silenced entry from configured Sensu instance
func (client *RestClient) UpdateSilenced(s *corev2.Silenced) error {
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}
	path := silencedPath(s.Namespace, s.Name)
	resp, err := client.R().SetBody(b).Put(path)
	if err != nil {
		return err
	}
	if resp.StatusCode() >= 400 {
		return UnmarshalError(resp)
	}
	return nil
}
