package v2

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

type resourceReferences struct {
	references []ResourceReference
	mu         sync.RWMutex
}

func (r *resourceReferences) add(ref ResourceReference) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.references = append(r.references, ref)
}

var (
	validPipelineWorkflowFilterReferences = resourceReferences{
		references: []ResourceReference{{APIVersion: "core/v2", Type: "EventFilter"}},
	}

	validPipelineWorkflowMutatorReferences = resourceReferences{
		references: []ResourceReference{{APIVersion: "core/v2", Type: "Mutator"}},
	}

	validPipelineWorkflowHandlerReferences = resourceReferences{
		references: []ResourceReference{{APIVersion: "core/v2", Type: "Handler"}},
	}
)

// AddValidPipelineWorkflowFilterReference adds a ResourceReference to the
// list of valid resource references for filters. Only the APIVersion and
// Type fields are used to validate resource references at this time.
func AddValidPipelineWorkflowFilterReference(ref ResourceReference) {
	validPipelineWorkflowFilterReferences.add(ref)
}

// AddValidPipelineWorkflowMutatorReference adds a ResourceReference to the
// list of valid resource references for mutators. Only the APIVersion and
// Type fields are used to validate resource references at this time.
func AddValidPipelineWorkflowMutatorReference(ref ResourceReference) {
	validPipelineWorkflowMutatorReferences.add(ref)
}

// AddValidPipelineWorkflowHandlerReference adds a ResourceReference to the
// list of valid resource references for handlers. Only the APIVersion and
// Type fields are used to validate resource references at this time.
func AddValidPipelineWorkflowHandlerReference(ref ResourceReference) {
	validPipelineWorkflowHandlerReferences.add(ref)
}

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
	for _, allowed := range validPipelineWorkflowFilterReferences.references {
		if allowed.APIVersion == ref.APIVersion && allowed.Type == ref.Type {
			return nil
		}
	}
	return fmt.Errorf("resource type not capable of filtering events: %s.%s", ref.APIVersion, ref.Type)
}

func (w *PipelineWorkflow) validateMutatorReference(ref *ResourceReference) error {
	for _, allowed := range validPipelineWorkflowMutatorReferences.references {
		if allowed.APIVersion == ref.APIVersion && allowed.Type == ref.Type {
			return nil
		}
	}
	return fmt.Errorf("resource type not capable of mutating events: %s.%s", ref.APIVersion, ref.Type)
}

func (w *PipelineWorkflow) validateHandlerReference(ref *ResourceReference) error {
	for _, allowed := range validPipelineWorkflowHandlerReferences.references {
		if allowed.APIVersion == ref.APIVersion && allowed.Type == ref.Type {
			return nil
		}
	}
	return fmt.Errorf("resource type not capable of handling events: %s.%s", ref.APIVersion, ref.Type)
}
