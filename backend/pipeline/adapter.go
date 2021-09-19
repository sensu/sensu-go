package pipeline

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// Adapter specifies an interface for pipeline adapters which are used to
// run references to pipelines.
type Adapter interface {
	Name() string
	CanRun(*corev2.ResourceReference) bool
	Run(context.Context, *corev2.ResourceReference, interface{}) error
}
