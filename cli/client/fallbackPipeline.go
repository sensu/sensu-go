package client

import (
	"encoding/json"
	"fmt"

	corev2 "github.com/sensu/core/v2"
)

// FallbackPipelinesPath is the api path for pipelines.
var FallbackPipelinesPath = createNSBasePath(coreAPIGroup, coreAPIVersion, "fallbackPipeline")

// FetchPipeline fetches a specific pipeline
func (client *RestClient) FetchFallbackPipeline(name string) (*corev2.FallbackPipeline, error) {
	var pipeline *corev2.FallbackPipeline
	fmt.Println("<<<<<<<<<<<<<<< GET FALLBACK PIPELINES")

	path := FallbackPipelinesPath(client.config.Namespace(), name)
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

// UpdatePipeline updates a pipeline.
func (client *RestClient) UpdateFallbackPipeline(pipeline *corev2.Pipeline) error {
	bytes, err := json.Marshal(pipeline)
	if err != nil {
		return err
	}

	path := FallbackPipelinesPath(pipeline.GetNamespace(), pipeline.GetName())
	res, err := client.R().SetBody(bytes).Put(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

//// Manisha creates fallbackpipeline
//func (client *RestClient) CreateFallbackPipeline(pipeline *corev2.FallbackPipeline) (err error) {
//
//	fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
//	fmt.Println("Manisha hits fallback ")
//	fmt.Println("++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
//	bytes, err := json.Marshal(pipeline)
//	if err != nil {
//		return err
//	}
//
//	path := EntitiesPath(pipeline.Namespace)
//	res, err := client.R().SetBody(bytes).Post(path)
//	if err != nil {
//		return err
//	}
//
//	if res.StatusCode() >= 400 {
//		return UnmarshalError(res)
//	}
//
//	return nil
//}
