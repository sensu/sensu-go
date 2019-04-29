package client

import (
	"encoding/json"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
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
func (client *RestClient) ListEvents(namespace string, options ListOptions) ([]corev2.Event, string, error) {
	var events []corev2.Event
	var header string

	path := eventsPath(namespace)
	request := client.R()

	ApplyListOptions(request, options)

	res, err := request.Get(path)
	if err != nil {
		return events, header, err
	}

	header = EntityLimitHeader(res.Header())

	if res.StatusCode() >= 400 {
		return nil, header, UnmarshalError(res)
	}

	err = json.Unmarshal(res.Body(), &events)
	return events, header, err
}

// DeleteEvent deletes an event.
func (client *RestClient) DeleteEvent(namespace, entity, check string) error {
	return client.Delete(eventsPath(namespace, entity, check))
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
