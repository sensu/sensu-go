package client

import (
	"encoding/json"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// EventsPath is the api path for events.
var EventsPath = createNSBasePath(coreAPIGroup, coreAPIVersion, "events")

// FetchEvent fetches a specific event
func (client *RestClient) FetchEvent(entity, check string) (*corev2.Event, error) {
	var event *corev2.Event

	path := EventsPath(client.config.Namespace(), entity, check)
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

// DeleteEvent deletes an event.
func (client *RestClient) DeleteEvent(namespace, entity, check string) error {
	return client.Delete(EventsPath(namespace, entity, check))
}

// UpdateEvent updates an event.
func (client *RestClient) UpdateEvent(event *corev2.Event) error {
	bytes, err := json.Marshal(event)
	if err != nil {
		return err
	}

	path := EventsPath(event.Check.Namespace, event.Entity.Name, event.Check.Name)
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
func (client *RestClient) ResolveEvent(event *corev2.Event) error {
	event.Check.Status = 0
	event.Check.Output = "Resolved manually by sensuctl"
	event.Timestamp = int64(time.Now().Unix())
	return client.UpdateEvent(event)
}
