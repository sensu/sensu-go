package v3

import (
	"errors"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// validate checks if a pipeline resource passes validation rules.
func (p *Pipeline) validate() error {
	if p.Metadata == nil {
		return errors.New("metadata must be set")
	}

	if err := corev2.ValidateName(p.Metadata.Name); err != nil {
		return errors.New("name " + err.Error())
	}

	if p.Metadata.Namespace == "" {
		return errors.New("namespace must be set")
	}

	for _, workflow := range p.Workflows {
		if err := workflow.validate(); err != nil {
			return fmt.Errorf("workflow %w", err)
		}
	}

	return nil
}

// validate checks if a pipeline workflow resource passes validation rules.
func (w *PipelineWorkflow) validate() error {
	if err := corev2.ValidateName(w.Name); err != nil {
		return errors.New("name " + err.Error())
	}

	if w.Filters != nil {
		for _, filter := range w.Filters {
			if err := filter.validate(); err != nil {
				return fmt.Errorf("filter %w", err)
			}
			if err := filter.ValidateEventFilterReference(); err != nil {
				return fmt.Errorf("filter %w", err)
			}
		}
	}

	if w.Mutator != nil {
		if err := w.Mutator.validate(); err != nil {
			return fmt.Errorf("mutator %w", err)
		}
		if err := w.Mutator.ValidateMutatorReference(); err != nil {
			return fmt.Errorf("mutator %w", err)
		}
	}

	if w.Handler == nil {
		return errors.New("handler must be set")
	}

	if err := w.Handler.validate(); err != nil {
		return fmt.Errorf("handler %w", err)
	}

	if err := w.Handler.ValidateHandlerReference(); err != nil {
		return fmt.Errorf("handler %w", err)
	}

	return nil
}

// validate checks if a resource reference resource passes validation rules.
func (r *ResourceReference) validate() error {
	if err := corev2.ValidateName(r.Name); err != nil {
		return errors.New("name " + err.Error())
	}

	if r.Type == "" {
		return errors.New("type must be set")
	}

	if r.APIVersion == "" {
		return errors.New("api_version must be set")
	}

	return nil
}

func (r *ResourceReference) ValidateEventFilterReference() error {
	switch r.APIVersion {
	case "core/v2":
		switch r.Type {
		case "EventFilter":
			return nil
		}
	}
	return fmt.Errorf("resource type not capable of filtering events: %s.%s", r.APIVersion, r.Type)
}

func (r *ResourceReference) ValidateMutatorReference() error {
	switch r.APIVersion {
	case "core/v2":
		switch r.Type {
		case "Mutator":
			return nil
		}
	}
	return fmt.Errorf("resource type not capable of mutating events: %s.%s", r.APIVersion, r.Type)
}

func (r *ResourceReference) ValidateHandlerReference() error {
	switch r.APIVersion {
	case "core/v2":
		switch r.Type {
		case "Handler":
			return nil
		}
	}
	return fmt.Errorf("resource type not capable of handling events: %s.%s", r.APIVersion, r.Type)
}
