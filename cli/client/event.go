package client

import (
	"encoding/json"
	"fmt"

	"github.com/sensu/sensu-go/types"
)

func eventPath(entity, check string) string {
	const path = "/events/%s/%s"
	return fmt.Sprintf(path, entity, check)
}

// FetchEvent fetches a specific event
func (client *RestClient) FetchEvent(entity, check string) (*types.Event, error) {
	var event *types.Event
	res, err := client.R().Get(eventPath(entity, check))
	if err != nil {
		return nil, err
	}

	if res.StatusCode() >= 400 {
		return nil, unmarshalError(res)
	}

	err = json.Unmarshal(res.Body(), &event)
	return event, err
}

// ListEvents fetches events from Sensu API
func (client *RestClient) ListEvents(org string) ([]types.Event, error) {
	var events []types.Event

	res, err := client.R().Get("/events?org=" + org)
	if err != nil {
		return events, err
	}

	if res.StatusCode() >= 400 {
		return nil, unmarshalError(res)
	}

	err = json.Unmarshal(res.Body(), &events)
	return events, err
}

// DeleteEvent deletes an event.
func (client *RestClient) DeleteEvent(entity, check string) error {
	res, err := client.R().Delete(eventPath(entity, check))
	if err != nil {
		return err
	}
	if res.StatusCode() >= 400 {
		return unmarshalError(res)
	}
	return nil
}
