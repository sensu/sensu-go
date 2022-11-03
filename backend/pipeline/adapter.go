package pipeline

import (
	"context"
	"fmt"

	corev2 "github.com/sensu/core/v2"
)

// Adapter specifies an interface for pipeline adapters which are used to
// run references to pipelines.
type Adapter interface {
	Name() string
	CanRun(*corev2.ResourceReference) bool
	Run(context.Context, *corev2.ResourceReference, interface{}) error
}

// ErrMisconfiguredPipeline interface implemented by errors that indicate
// a misconfigured pipeline that cannot be executed.
type ErrMisconfiguredPipeline interface {
	MisconfiguredPipeline()
}

// ErrNoWorkflows is returned when a pipeline has no workflows
type ErrNoWorkflows struct{}

func (e *ErrNoWorkflows) Error() string {
	return "pipeline has no workflows"
}

func (e *ErrNoWorkflows) MisconfiguredPipeline() {}

// ErrNoLegacyHandlers is returned when no legacy handlers exist
type errNoLegacyHandlers struct {
	Msg string
}

func (e *errNoLegacyHandlers) Error() string {
	return fmt.Sprintf("legacy pipeline found no existing handlers: %s", e.Msg)
}

func (e *errNoLegacyHandlers) MisconfiguredPipeline() {}
