package client

import (
	"encoding/json"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

const healthPath = "/health"

// Health returns the health of the cluster.
func (c *RestClient) Health() (*corev2.HealthResponse, error) {
	res, err := c.R().Get(healthPath + "?timeout=3")
	if err != nil {
		return nil, fmt.Errorf("GET %q: %s", healthPath, err)
	}
	var healthResponse *corev2.HealthResponse
	return healthResponse, json.Unmarshal(res.Body(), &healthResponse)
}
