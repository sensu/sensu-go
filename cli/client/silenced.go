package client

import (
	"encoding/json"

	"github.com/sensu/sensu-go/types"
)

var silencedPath = createNSBasePath(coreAPIGroup, coreAPIVersion, "silenced")

// CreateSilenced creates a new silenced  entry.
func (client *RestClient) CreateSilenced(silenced *types.Silenced) error {
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
func (client *RestClient) DeleteSilenced(id string) error {
	path := silencedPath(client.config.Namespace(), id)
	res, err := client.R().Delete(path)
	if err != nil {
		return err
	}
	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}
	return nil
}

// ListSilenceds fetches all silenced entries from configured Sensu instance
func (client *RestClient) ListSilenceds(namespace, sub, check string) ([]types.Silenced, error) {
	if sub != "" && check != "" {
		name, err := types.SilencedName(sub, check)
		if err != nil {
			return nil, err
		}
		silenced, err := client.FetchSilenced(name)
		if err != nil {
			return nil, err
		}
		return []types.Silenced{*silenced}, nil
	}
	path := silencedPath(namespace)
	if sub != "" {
		path = silencedPath(namespace, "subscriptions", sub)
	} else if check != "" {
		path = silencedPath(namespace, "checks", check)
	}
	resp, err := client.R().Get(path)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return nil, UnmarshalError(resp)
	}

	var result []types.Silenced
	err = json.Unmarshal(resp.Body(), &result)
	return result, err
}

// FetchSilenced fetches a silenced entry from configured Sensu instance
func (client *RestClient) FetchSilenced(name string) (*types.Silenced, error) {
	path := silencedPath(client.config.Namespace(), name)
	resp, err := client.R().Get(path)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode() >= 400 {
		return nil, UnmarshalError(resp)
	}
	var result types.Silenced
	return &result, json.Unmarshal(resp.Body(), &result)
}

// UpdateSilenced updates a silenced entry from configured Sensu instance
func (client *RestClient) UpdateSilenced(s *types.Silenced) error {
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
