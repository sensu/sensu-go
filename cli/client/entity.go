package client

// ListEntities fetches all entities from configured Sensu instance
func (client *RestClient) ListEntities() (entities []types.Entity, err error) {
	r, err := client.R().Get("/entities")
	if err == nil {
		err = json.Unmarshal(r.Body(), &entities)
	}

	return
}
