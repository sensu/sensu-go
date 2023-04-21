package client

import (
	"encoding/json"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/core/v3/types"
)

// EntitiesPath is the api path for entities.
var EntitiesPath = createNSBasePath(coreAPIGroup, coreAPIVersion, "entities")

// DeleteEntity deletes given entitiy from the configured sensu instance
func (client *RestClient) DeleteEntity(namespace, name string) (err error) {
	return client.Delete(EntitiesPath(namespace, name))
}

// FetchEntity fetches a specific entity
func (client *RestClient) FetchEntity(name string) (*corev2.Entity, error) {
	path := EntitiesPath(client.config.Namespace(), name)
	res, err := client.R().Get(path)
	if err != nil {
		return nil, err
	}

	if res.StatusCode() >= 400 {
		return nil, UnmarshalError(res)
	}

	var warpper types.Wrapper
	err = json.Unmarshal(res.Body(), &warpper)
	return warpper.Value.(*corev2.Entity), err
}

// UpdateEntity updates given entity on configured Sensu instance
func (client *RestClient) UpdateEntity(entity *corev2.Entity) (err error) {
	bytes, err := json.Marshal(types.WrapResource(entity))
	if err != nil {
		return err
	}

	path := EntitiesPath(entity.Namespace, entity.Name)
	res, err := client.R().SetBody(bytes).Put(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// CreateEntity creates a new entity
func (client *RestClient) CreateEntity(entity *corev2.Entity) (err error) {
	bytes, err := json.Marshal(types.WrapResource(entity))
	if err != nil {
		return err
	}

	path := EntitiesPath(entity.Namespace)
	res, err := client.R().SetBody(bytes).Post(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}
