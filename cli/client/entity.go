package client

import (
	"encoding/json"

	"github.com/sensu/sensu-go/types"
)

// ListEntities fetches all entities from configured Sensu instance
func (client *RestClient) ListEntities() (entities []types.Entity, err error) {
	r, err := client.R().Get("/entities")
	if err == nil {
		err = json.Unmarshal(r.Body(), &entities)
	}

	return
}

// FetchEntity fetches all entities from configured Sensu instance
func (client *RestClient) FetchEntity(ID string) (entity types.Entity, err error) {
	r, err := client.R().Get("/entities/" + ID)
	if err == nil {
		err = json.Unmarshal(r.Body(), &entity)
	}

	return
}

// DeleteEntity deletes given entitiy from the configured sensu instance
func (client *RestClient) DeleteEntity(entity *types.Entity) (err error) {
	_, err = client.R().Delete("/entities/" + entity.ID)
	return err
}
