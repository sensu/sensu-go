package v2

import (
	"errors"
	"fmt"
)

// validate checks if a pipeline resource passes validation rules.
func (p *Pipeline) Validate() error {
	if p.Metadata == nil {
		return errors.New("metadata must be set")
	}

	if err := ValidateName(p.Metadata.Name); err != nil {
		return errors.New("name " + err.Error())
	}

	if p.Metadata.Namespace == "" {
		return errors.New("namespace must be set")
	}

	for _, workflow := range p.Workflows {
		if err := workflow.Validate(); err != nil {
			return fmt.Errorf("workflow %w", err)
		}
	}

	return nil
}
