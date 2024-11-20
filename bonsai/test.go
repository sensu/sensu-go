package v2

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"sync"
	"time"

	stringsutil "github.com/sensu/core/v2/internal/stringutil"
)

const (
	// FallbackPipelinesResource is the name of this resource type
	FallbackPipelineResource = "fallbackPipeline"
)

// FallbackPipeline defines a list of pipelines with independent execution.
type FallbackPipeline struct {
	ObjectMeta   ObjectMeta
	Namespace    string
	PipelineList []*ResourceReference // List of pipelines to execute
}

// ExecuteFallbackPipeline executes each pipeline independently, logging errors if any fail.
func (f *FallbackPipeline) ExecuteFallbackPipeline() {
	var wg sync.WaitGroup

	for _, pipelineRef := range f.PipelineList {
		wg.Add(1)
		go func(pipeline *ResourceReference) {
			defer wg.Done()
			if err := executePipeline(pipeline); err != nil {
				fmt.Printf("Pipeline %s failed: %v\n", pipeline.Name, err)
			} else {
				fmt.Printf("Pipeline %s succeeded.\n", pipeline.Name)
			}
		}(pipelineRef)
	}

	wg.Wait() // Wait for all goroutines to finish
}

// executePipeline simulates executing a single pipeline.
// Replace this with actual execution logic.
func executePipeline(pipeline *ResourceReference) error {
	fmt.Printf("Executing pipeline: %s\n", pipeline.Name)
	// Simulate pipeline execution time
	time.Sleep(1 * time.Second)
	// Simulate a random failure
	if pipeline.Name == "fail" {
		return errors.New("simulated pipeline failure")
	}
	return nil
}

// GetObjectMeta returns the object metadata for the resource.
func (f *FallbackPipeline) GetObjectMeta() ObjectMeta {
	return f.ObjectMeta
}

// SetObjectMeta sets the object metadata for the resource.
func (f *FallbackPipeline) SetObjectMeta(meta ObjectMeta) {
	f.ObjectMeta = meta
}

// SetNamespace sets the namespace of the resource.
func (f *FallbackPipeline) SetNamespace(namespace string) {
	f.Namespace = namespace
}

// StorePrefix returns the path prefix to this resource in the store.
func (f *FallbackPipeline) StorePrefix() string {
	return FallbackPipelineResource
}

// RBACName describes the name of the resource for RBAC purposes.
func (f *FallbackPipeline) RBACName() string {
	return FallbackPipelineResource
}

// URIPath gives the path component of a pipeline URI.
func (f *FallbackPipeline) URIPath() string {
	if f.Namespace == "" {
		return path.Join(URLPrefix, FallbackPipelineResource, url.PathEscape(f.Name))
	}
	return path.Join(URLPrefix, "namespaces", url.PathEscape(f.Namespace), FallbackPipelineResource, url.PathEscape(f.Name))
}

// Validate checks if a pipeline resource passes validation rules.
func (f *FallbackPipeline) Validate() error {
	if err := ValidateName(f.ObjectMeta.Name); err != nil {
		return errors.New("name " + err.Error())
	}
	if f.ObjectMeta.Namespace == "" {
		return errors.New("namespace must be set")
	}
	return nil
}

// FallbackPipelineFields returns a set of fields that represent that resource.
func FallbackPipelineFields(r Resource) map[string]string {
	resource := r.(*FallbackPipeline)
	fields := map[string]string{
		"fallbackPipeline.name":      resource.ObjectMeta.Name,
		"fallbackPipeline.namespace": resource.ObjectMeta.Namespace,
	}
	stringsutil.MergeMapWithPrefix(fields, resource.ObjectMeta.Labels, "fallbackPipeline.labels.")
	return fields
}

// Fields returns a set of fields that represent that resource.
func (f *FallbackPipeline) Fields() map[string]string {
	return FallbackPipelineFields(f)
}

// FixtureFallbackPipeline returns a testing fixture for a FallbackPipeline object.
func FixtureFallbackPipeline(name, namespace string) *FallbackPipeline {
	return &FallbackPipeline{
		ObjectMeta:   NewObjectMeta(name, namespace),
		PipelineList: []*ResourceReference{},
	}
}

// FixtureFallbackPipelineReference returns a testing fixture for a ResourceReference
// object referencing a corev2.Pipeline.
func FixtureFallbackPipelineReference(name string) *ResourceReference {
	return &ResourceReference{
		APIVersion: "core/v2",
		Type:       "FallbackPipeline",
		Name:       name,
	}
}
