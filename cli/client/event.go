package client

import (
	"encoding/json"
	"time"

	"github.com/sensu/sensu-go/types"
)

var eventsPath = createNSBasePath(coreAPIGroup, coreAPIVersion, "events")

// FetchEvent fetches a specific event
func (client *RestClient) FetchEvent(entity, check string) (*types.Event, error) {
	var event *types.Event

	path := eventsPath(client.config.Namespace(), entity, check)
	res, err := client.R().Get(path)
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
func (client *RestClient) ListEvents(namespace string) ([]types.Event, error) {
	var events []types.Event

	path := eventsPath(namespace)
	res, err := client.R().Get(path)
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
	path := eventsPath(client.config.Namespace(), entity, check)
	res, err := client.R().Delete(path)
	if err != nil {
		return err
	}
	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}
	return nil
}

// UpdateEvent updates an event.
func (client *RestClient) UpdateEvent(event *types.Event) error {
	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	path := eventsPath(event.Check.Namespace, event.Entity.Name, event.Check.Name)
	res, err := client.R().SetBody(bytes).Put(path)
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
	return client.UpdateEvent(event)
}
