package client

import (
	"encoding/json"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/types"
)

// HandlersPath is the api path for handlers.
var HandlersPath = createNSBasePath(coreAPIGroup, coreAPIVersion, "handlers")

// CreateHandler creates new handler on configured Sensu instance
func (client *RestClient) CreateHandler(handler *corev2.Handler) (err error) {
	bytes, err := json.Marshal(types.WrapResource(handler))
	if err != nil {
		return err
	}

	path := HandlersPath(handler.Namespace)
	res, err := client.R().SetBody(bytes).Post(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// DeleteHandler deletes given handler from the configured Sensu instance
func (client *RestClient) DeleteHandler(namespace, name string) (err error) {
	return client.Delete(HandlersPath(namespace, name))
}

// FetchHandler fetches a specific handler
func (client *RestClient) FetchHandler(name string) (*corev2.Handler, error) {
	path := HandlersPath(client.config.Namespace(), name)
	res, err := client.R().Get(path)
	if err != nil {
		return nil, err
	}

	if res.StatusCode() >= 400 {
		return nil, UnmarshalError(res)
	}

	var wrapper types.Wrapper
	err = json.Unmarshal(res.Body(), &wrapper)
	return wrapper.Value.(*corev2.Handler), err
}

// UpdateHandler updates given handler on configured Sensu instance
func (client *RestClient) UpdateHandler(handler *corev2.Handler) (err error) {
	bytes, err := json.Marshal(types.WrapResource(handler))
	if err != nil {
		return err
	}

	path := HandlersPath(handler.Namespace, handler.Name)
	res, err := client.R().SetBody(bytes).Put(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}
