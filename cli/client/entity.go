package client

import (
	"encoding/json"

	"github.com/sensu/sensu-go/types"
)

// DeleteEntity deletes given entitiy from the configured sensu instance
func (client *RestClient) DeleteEntity(entity *types.Entity) (err error) {
	_, err = client.R().Delete("/entities/" + entity.Name)
	return err
}

// FetchEntity fetches a specific entity
func (client *RestClient) FetchEntity(name string) (*types.Entity, error) {
	var entity *types.Entity
	res, err := client.R().Get("/entities/" + name)
	if err != nil {
		return entity, err
	}

	if res.StatusCode() >= 400 {
		return entity, UnmarshalError(res)
	}

	err = json.Unmarshal(res.Body(), &entity)
	return entity, err
}

// ListEntities fetches all entities from configured Sensu instance
func (client *RestClient) ListEntities(namespace string) ([]types.Entity, error) {
	var entities []types.Entity

	res, err := client.R().SetQueryParam("namespace", namespace).Get("/entities")
	if err != nil {
		return entities, err
	}

	if res.StatusCode() >= 400 {
		return entities, UnmarshalError(res)
	}

	err = json.Unmarshal(res.Body(), &entities)
	return entities, err
}

// UpdateEntity updates given entity on configured Sensu instance
func (client *RestClient) UpdateEntity(entity *types.Entity) (err error) {
	bytes, err := json.Marshal(entity)
	if err != nil {
		return err
	}

	res, err := client.R().SetBody(bytes).Put("/entities/" + entity.Name)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// CreateEntity creates a new entity
func (client *RestClient) CreateEntity(entity *types.Entity) (err error) {
	bytes, err := json.Marshal(entity)
	if err != nil {
		return err
	}

	res, err := client.R().SetBody(bytes).Post("/entities")
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}
