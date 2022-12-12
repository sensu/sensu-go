package client

import (
	"encoding/json"

	corev2 "github.com/sensu/core/v2"
)

// PipelinesPath is the api path for pipelines.
var PipelinesPath = createNSBasePath(coreAPIGroup, coreAPIVersion, "pipelines")

// FetchPipeline fetches a specific pipeline
func (client *RestClient) FetchPipeline(name string) (*corev2.Pipeline, error) {
	var pipeline *corev2.Pipeline

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

// DeletePipeline deletes a pipeline.
func (client *RestClient) DeletePipeline(namespace, name string) error {
	return client.Delete(PipelinesPath(namespace, name))
}

// UpdatePipeline updates a pipeline.
func (client *RestClient) UpdatePipeline(pipeline *corev2.Pipeline) error {
	bytes, err := json.Marshal(pipeline)
	if err != nil {
		return err
	}

	path := PipelinesPath(pipeline.GetNamespace(), pipeline.GetName())
	res, err := client.R().SetBody(bytes).Put(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// CreatePipeline creates a new pipeline
func (client *RestClient) CreatePipeline(pipeline *corev2.Pipeline) (err error) {
	bytes, err := json.Marshal(pipeline)
	if err != nil {
		return err
	}

	path := EntitiesPath(pipeline.Namespace)
	res, err := client.R().SetBody(bytes).Post(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}
