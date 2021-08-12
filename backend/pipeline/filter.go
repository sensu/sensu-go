package pipeline

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

type Filter interface {
	CanFilter(context.Context, *corev2.ResourceReference) bool
	Filter(context.Context, *corev2.ResourceReference, *corev2.Event) (bool, error)
}
