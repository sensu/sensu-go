package client

import (
	"encoding/json"
	"fmt"

	"github.com/sensu/sensu-go/types"
)

// FetchEvent fetches a specific event
func (client *RestClient) FetchEvent(entity, check string) (*types.Event, error) {
	var event *types.Event
	res, err := client.R().Get("/events/" + entity + "/" + check)
	if err != nil {
		return nil, err
	}

	if res.StatusCode() >= 400 {
		return nil, fmt.Errorf("%v", res.String())
	}

	err = json.Unmarshal(res.Body(), &event)
	return event, err
}

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
