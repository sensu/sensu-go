package client

import (
	"encoding/json"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/types"
)

var entitiesPath = createNSBasePath(coreAPIGroup, coreAPIVersion, "entities")

// DeleteEntity deletes given entitiy from the configured sensu instance
func (client *RestClient) DeleteEntity(namespace, name string) (err error) {
	return client.Delete(entitiesPath(namespace, name))
}

// FetchEntity fetches a specific entity
func (client *RestClient) FetchEntity(name string) (*types.Entity, error) {
	var entity *types.Entity

	path := entitiesPath(client.config.Namespace(), name)
	res, err := client.R().Get(path)
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
func (client *RestClient) ListEntities(namespace string, options ListOptions) ([]corev2.Entity, string, error) {
	var entities []corev2.Entity
	var header string

	path := entitiesPath(namespace)
	request := client.R()

	ApplyListOptions(request, options)

	res, err := request.Get(path)
	if err != nil {
		return entities, header, err
	}

	header = EntityLimitHeader(res.Header())

	if res.StatusCode() >= 400 {
		return entities, header, UnmarshalError(res)
	}

	err = json.Unmarshal(res.Body(), &entities)
	return entities, header, err
}

// UpdateEntity updates given entity on configured Sensu instance
func (client *RestClient) UpdateEntity(entity *types.Entity) (err error) {
	bytes, err := json.Marshal(entity)
	if err != nil {
		return err
	}

	path := entitiesPath(entity.Namespace, entity.Name)
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
func (client *RestClient) CreateEntity(entity *types.Entity) (err error) {
	bytes, err := json.Marshal(entity)
	if err != nil {
		return err
	}

	path := entitiesPath(entity.Namespace)
	res, err := client.R().SetBody(bytes).Post(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}
