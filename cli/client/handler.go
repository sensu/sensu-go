package client

import (
	"encoding/json"
	"fmt"

	"github.com/sensu/sensu-go/types"
)

var handlersPath = createNSBasePath(coreAPIGroup, coreAPIVersion, "handlers")

// ListHandlers fetches all handlers from configured Sensu instance
func (client *RestClient) ListHandlers(namespace string) ([]types.Handler, error) {
	var handlers []types.Handler

	path := handlersPath(namespace)
	res, err := client.R().Get(path)
	if err != nil {
		return handlers, err
	}

	if res.StatusCode() >= 400 {
		return handlers, UnmarshalError(res)
	}

	err = json.Unmarshal(res.Body(), &handlers)
	return handlers, err
}

// CreateHandler creates new handler on configured Sensu instance
func (client *RestClient) CreateHandler(handler *types.Handler) (err error) {
	bytes, err := json.Marshal(handler)
	if err != nil {
		return err
	}

	path := handlersPath(handler.Namespace)
	res, err := client.R().SetBody(bytes).Post(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return fmt.Errorf("%v", res.String())
	}

	return nil
}

// DeleteHandler deletes given handler from the configured Sensu instance
func (client *RestClient) DeleteHandler(handler *types.Handler) (err error) {
	path := handlersPath(client.config.Namespace(), handler.Name)
	res, err := client.R().Delete(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return fmt.Errorf("%v", res.String())
	}

	return nil
}

// FetchHandler fetches a specific handler
func (client *RestClient) FetchHandler(name string) (*types.Handler, error) {
	var handler *types.Handler
	path := handlersPath(client.config.Namespace(), name)
	res, err := client.R().Get(path)
	if err != nil {
		return nil, err
	}

	if res.StatusCode() >= 400 {
		return nil, UnmarshalError(res)
	}

	err = json.Unmarshal(res.Body(), &handler)
	return handler, err
}

// UpdateHandler updates given handler on configured Sensu instance
func (client *RestClient) UpdateHandler(handler *types.Handler) (err error) {
	bytes, err := json.Marshal(handler)
	if err != nil {
		return err
	}

	path := handlersPath(handler.Namespace, handler.Name)
	res, err := client.R().SetBody(bytes).Put(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return fmt.Errorf("%v", res.String())
	}

	return nil
}
