package v2

import (
	"errors"
	"fmt"
	"net/url"
	"path"
)

const (
	// PipelinesResource is the name of this resource type
	PipelinesResource = "pipelines"
)

// GetObjectMeta returns the object metadata for the resource.
func (p Pipeline) GetObjectMeta() ObjectMeta {
	return p.ObjectMeta
}

// SetObjectMeta sets the object metadata for the resource.
func (p Pipeline) SetObjectMeta(meta ObjectMeta) {
	p.ObjectMeta = meta
}

// SetNamespace sets the namespace of the resource.
func (p Pipeline) SetNamespace(namespace string) {
	p.Namespace = namespace
}

// StorePrefix returns the path prefix to this resource in the store.
func (p *Pipeline) StorePrefix() string {
	return PipelinesResource
}

// RBACName describes the name of the resource for RBAC purposes.
func (p *Pipeline) RBACName() string {
	return "pipelines"
}

// URIPath gives the path component of a pipeline URI.
func (p *Pipeline) URIPath() string {
	return path.Join(URLPrefix, "namespaces", url.PathEscape(p.Namespace), PipelinesResource, url.PathEscape(p.Name))
}

// Validate checks if a pipeline resource passes validation rules.
func (p *Pipeline) Validate() error {
	if err := ValidateName(p.ObjectMeta.Name); err != nil {
		return errors.New("name " + err.Error())
	}

	if p.ObjectMeta.Namespace == "" {
		return errors.New("namespace must be set")
	}

	for _, workflow := range p.Workflows {
		if err := workflow.Validate(); err != nil {
			return fmt.Errorf("workflow %w", err)
		}
	}

	return nil
}

// FixturePipelineReference returns a testing fixture for a ResourceReference
// object referencing a corev2.Pipeline.
func FixturePipelineReference(name string) *ResourceReference {
	return &ResourceReference{
		APIVersion: "core/v2",
		Type:       "Pipeline",
		Name:       name,
	}
}
