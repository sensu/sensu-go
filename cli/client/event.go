package client

import (
	"encoding/json"

	"github.com/sensu/sensu-go/types"
)

// ListEvents fetches events from Sensu API
func (client *RestClient) ListEvents() (events []types.Event, err error) {
	r, err := client.R().Get("/events")
	if err == nil {
		err = json.Unmarshal(r.Body(), &events)
	}

	return
}
