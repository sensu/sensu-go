package v2

import (
	"errors"
	"fmt"
)

// validate checks if a pipeline resource passes validation rules.
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
