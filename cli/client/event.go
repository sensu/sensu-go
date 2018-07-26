package client

import (
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"github.com/sensu/sensu-go/types"
)

func eventPath(entity, check string) string {
	const path = "/events/%s/%s"
	return fmt.Sprintf(path, url.PathEscape(entity), url.PathEscape(check))
}

// FetchEvent fetches a specific event
func (client *RestClient) FetchEvent(entity, check string) (*types.Event, error) {
	var event *types.Event
	res, err := client.R().Get(eventPath(entity, check))
	if err != nil {
		return nil, err
	}

	if res.StatusCode() >= 400 {
		return nil, UnmarshalError(res)
	}

	err = json.Unmarshal(res.Body(), &event)
	return event, err
}

// ListEvents fetches events from Sensu API
func (client *RestClient) ListEvents(org string) ([]types.Event, error) {
	var events []types.Event

	res, err := client.R().Get("/events?org=" + url.QueryEscape(org))
	if err != nil {
		return events, err
	}

	if res.StatusCode() >= 400 {
		return nil, UnmarshalError(res)
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
		return UnmarshalError(res)
	}
	return nil
}

// ResolveEvent resolves an event.
func (client *RestClient) ResolveEvent(event *types.Event) error {
	event.Check.Status = 0
	event.Check.Output = "Resolved manually by sensuctl"
	event.Timestamp = int64(time.Now().Unix())

	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	res, err := client.R().SetBody(bytes).Put("/events/" + event.Entity.ID + "/" + event.Check.Name)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}
