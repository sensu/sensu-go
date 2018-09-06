package client

import (
	"encoding/json"
	"fmt"

	"github.com/sensu/sensu-go/types"
)

const healthPath = "/health"

func (c *RestClient) Health() (*types.HealthResponse, error) {
	res, err := c.R().Get(healthPath)
	if err != nil {
		return nil, fmt.Errorf("GET %q: %s", healthPath, err)
	}
	var healthResponse *types.HealthResponse
	return healthResponse, json.Unmarshal(res.Body(), &healthResponse)
}
