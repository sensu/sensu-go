package client

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/sensu/sensu-go/types"
)

// ListHandlers fetches all handlers from configured Sensu instance
func (client *RestClient) ListHandlers(org string) ([]types.Handler, error) {
	var handlers []types.Handler

	res, err := client.R().Get("/handlers?org=" + url.QueryEscape(org))
	if err != nil {
		return handlers, err
	}

	if res.StatusCode() >= 400 {
		return handlers, fmt.Errorf("%v", res.String())
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

	res, err := client.R().SetBody(bytes).Post("/handlers")
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
	res, err := client.R().Delete("/handlers/" + url.PathEscape(handler.Name))
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
	res, err := client.R().Get("/handlers/" + url.PathEscape(name))
	if err != nil {
		return nil, err
	}

	if res.StatusCode() >= 400 {
		return nil, fmt.Errorf("%v", res.String())
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

	res, err := client.R().SetBody(bytes).Put("/handlers/" + url.PathEscape(handler.Name))
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return fmt.Errorf("%v", res.String())
	}

	return nil
}
