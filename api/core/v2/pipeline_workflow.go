package v2

import (
	"context"
	"errors"
	"fmt"
)

// PipelineWorkflowFromHandler takes a Handler, converts it to a
// PipelineWorkflow and then returns it.
func PipelineWorkflowFromHandler(ctx context.Context, workflowName string, handler *Handler) *PipelineWorkflow {
	filterRefs := []*ResourceReference{}
	for _, filterName := range handler.Filters {
		ref := &ResourceReference{
			Name:       filterName,
			APIVersion: "core/v2",
			Type:       "EventFilter",
		}
		filterRefs = append(filterRefs, ref)
	}

	var mutatorRef *ResourceReference
	if handler.Mutator != "" {
		mutatorRef = &ResourceReference{
			Name:       handler.Mutator,
			APIVersion: "core/v2",
			Type:       "Mutator",
		}
	}

	handlerRef := &ResourceReference{
		Name:       handler.Name,
		APIVersion: "core/v2",
		Type:       "Handler",
	}

	return &PipelineWorkflow{
		Name:    workflowName,
		Filters: filterRefs,
		Mutator: mutatorRef,
		Handler: handlerRef,
	}
}

// validate checks if a pipeline workflow resource passes validation rules.
func (w *PipelineWorkflow) Validate() error {
	if err := ValidateName(w.Name); err != nil {
		return errors.New("name " + err.Error())
	}

	if w.Filters != nil {
		for _, filter := range w.Filters {
			if err := filter.Validate(); err != nil {
				return fmt.Errorf("filter %w", err)
			}
			if err := w.validateEventFilterReference(filter); err != nil {
				return fmt.Errorf("filter %w", err)
			}
		}
	}

	if w.Mutator != nil {
		if err := w.Mutator.Validate(); err != nil {
			return fmt.Errorf("mutator %w", err)
		}
		if err := w.validateMutatorReference(w.Mutator); err != nil {
			return fmt.Errorf("mutator %w", err)
		}
	}

	if w.Handler == nil {
		return errors.New("handler must be set")
	}

	if err := w.Handler.Validate(); err != nil {
		return fmt.Errorf("handler %w", err)
	}

	if err := w.validateHandlerReference(w.Handler); err != nil {
		return fmt.Errorf("handler %w", err)
	}

	return nil
}

func (w *PipelineWorkflow) validateEventFilterReference(ref *ResourceReference) error {
	switch ref.APIVersion {
	case "core/v2":
		switch ref.Type {
		case "EventFilter":
			return nil
		}
	}
	return fmt.Errorf("resource type not capable of filtering events: %s.%s", ref.APIVersion, ref.Type)
}

func (w *PipelineWorkflow) validateMutatorReference(ref *ResourceReference) error {
	switch ref.APIVersion {
	case "core/v2":
		switch ref.Type {
		case "Mutator":
			return nil
		}
	}
	return fmt.Errorf("resource type not capable of mutating events: %s.%s", ref.APIVersion, ref.Type)
}

func (w *PipelineWorkflow) validateHandlerReference(ref *ResourceReference) error {
	switch ref.APIVersion {
	case "core/v2":
		switch ref.Type {
		case "Handler":
			return nil
		}
	}
	return fmt.Errorf("resource type not capable of handling events: %s.%s", ref.APIVersion, ref.Type)
}
