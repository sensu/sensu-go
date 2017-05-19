package client

import (
	"encoding/json"
	"fmt"

	"github.com/sensu/sensu-go/types"
)

// ListEvents fetches events from Sensu API
func (client *RestClient) ListEvents() ([]types.Event, error) {
	var events []types.Event

	res, err := client.R().Get("/events")
	if err != nil {
		return events, err
	}

	if res.StatusCode() >= 400 {
		return events, fmt.Errorf("%v", res.String())
	}

	err = json.Unmarshal(res.Body(), &events)
	return events, err
}
