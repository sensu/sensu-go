package pipeline

import (
	"context"

	corev2 "github.com/sensu/core/v2"
)

// Adapter specifies an interface for pipeline adapters which are used to
// run references to pipelines.
type Adapter interface {
	Name() string
	CanRun(*corev2.ResourceReference) bool
	Run(context.Context, *corev2.ResourceReference, interface{}) error
}

// ErrNoWorkflows is returned when a pipeline has no workflows
type ErrNoWorkflows struct{}

func (e *ErrNoWorkflows) Error() string {
	return "pipeline has no workflows"
}
