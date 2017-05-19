package client

import (
	"encoding/json"
	"fmt"

	"github.com/sensu/sensu-go/types"
)

// ListEntities fetches all entities from configured Sensu instance
func (client *RestClient) ListEntities() ([]types.Entity, error) {
	var entities []types.Entity

	res, err := client.R().Get("/entities")
	if err != nil {
		return entities, err
	}

	if res.StatusCode() >= 400 {
		return entities, fmt.Errorf("%v", res.String())
	}

	err = json.Unmarshal(res.Body(), &entities)
	return entities, err
}

// FetchEntity fetches all entities from configured Sensu instance
func (client *RestClient) FetchEntity(ID string) (types.Entity, error) {
	var entity types.Entity
	res, err := client.R().Get("/entities/" + ID)
	if err != nil {
		return entity, err
	}

	if res.StatusCode() >= 400 {
		return entity, fmt.Errorf("%v", res.String())
	}

	err = json.Unmarshal(res.Body(), &entity)
	return entity, err
}
