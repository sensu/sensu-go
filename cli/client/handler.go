package client

import (
	"encoding/json"

	"github.com/sensu/sensu-go/types"
)

// ListChecks fetches all handlers from configured Sensu instance
func (client *RestClient) ListHandlers() (handlers []types.Handler, err error) {
	r, err := client.R().Get("/handlers")
	if err == nil {
		err = json.Unmarshal(r.Body(), &handlers)
	}

	return
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
