package client

import (
	"encoding/json"
	"time"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/types"
)

// EventsPath is the api path for events.
var EventsPath = createNSBasePath(coreAPIGroup, coreAPIVersion, "events")

// FetchEvent fetches a specific event
func (client *RestClient) FetchEvent(entity, check string) (*corev2.Event, error) {
	path := EventsPath(client.config.Namespace(), entity, check)
	res, err := client.R().Get(path)
	if err != nil {
		return nil, err
	}

	if res.StatusCode() >= 400 {
		return nil, UnmarshalError(res)
	}

	var wrapper types.Wrapper
	err = json.Unmarshal(res.Body(), &wrapper)
	return wrapper.Value.(*corev2.Event), err
}

// DeleteEvent deletes an event.
func (client *RestClient) DeleteEvent(namespace, entity, check string) error {
	return client.Delete(EventsPath(namespace, entity, check))
}

// UpdateEvent updates an event.
func (client *RestClient) UpdateEvent(event *corev2.Event) error {
	bytes, err := json.Marshal(types.WrapResource(event))
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
	event.Check.Executed = int64(time.Now().Unix())
	event.Timestamp = event.Check.Executed
	return client.UpdateEvent(event)
}
