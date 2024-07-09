package client

import (
	"encoding/json"
	"fmt"

	corev2 "github.com/sensu/core/v2"
)

// FallbackPipelinesPath is the api path for pipelines.
var FallbackPipelinesPath = createNSBasePath(coreAPIGroup, coreAPIVersion, "fallbackPipeline")

// FetchPipeline fetches a specific pipeline
func (client *RestClient) FetchFallbackPipeline(name string) (*corev2.Pipeline, error) {
	var pipeline *corev2.Pipeline
	fmt.Println("<<<<<<<<<<<<<<< GET PIPELINES")
	fmt.Println("<<<<<<<<<<<<<<<")

	path := PipelinesPath(client.config.Namespace(), name)
	res, err := client.R().Get(path)
	if err != nil {
		return nil, err
	}

	if res.StatusCode() >= 400 {
		return nil, UnmarshalError(res)
	}

	err = json.Unmarshal(res.Body(), &pipeline)
	return pipeline, err
}

// DeleteFallbackPipeline deletes a fallbackpipeline.
func (client *RestClient) DeleteFallbackPipeline(namespace, name string) error {
	return client.Delete(FallbackPipelinesPath(namespace, name))
}
