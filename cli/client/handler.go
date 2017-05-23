package client

import (
	"encoding/json"
	"fmt"

	"github.com/sensu/sensu-go/types"
)

// ListHandlers fetches all handlers from configured Sensu instance
func (client *RestClient) ListHandlers() ([]types.Handler, error) {
	var handlers []types.Handler

	res, err := client.R().Get("/handlers")
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
	if err == nil {
		_, err = client.R().
			SetBody(bytes).
			Put("/handlers/" + handler.Name)
	}

	return
}

// DeleteHandler deletes given handler from the configured Sensu instance
func (client *RestClient) DeleteHandler(handler *types.Handler) (err error) {
	_, err = client.R().Delete("/handlers/" + handler.Name)
	return err
}
