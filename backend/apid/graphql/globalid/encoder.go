package globalid

import (
	"context"

	corev2 "github.com/sensu/core/v2"
)

type EncodeFn func(context.Context, interface{}) *StandardComponents

// DefaultEncoder is the default implementation of Encode.
func DefaultEncoder(ctx context.Context, res interface{}) *StandardComponents {
	if res, ok := res.(interface{ GetMetadata() *corev2.ObjectMeta }); ok {
		return DefaultEncoder(ctx, res.GetMetadata())
	}

	cmp := StandardComponents{}
	if res, ok := res.(interface{ GetName() string }); ok {
		cmp.uniqueComponent = res.GetName()
	}
	if res, ok := res.(interface{ GetNamespace() string }); ok {
		cmp.namespace = res.GetNamespace()
	}
	return &cmp
}

// Encode produces a globalid components for a given resource.
var Encode EncodeFn = DefaultEncoder
