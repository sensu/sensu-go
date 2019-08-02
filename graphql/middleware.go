package graphql

import (
	"context"

	"github.com/graphql-go/graphql"
)

// Middleware is an interface for extending a service with operations that
// wrap existing functionality.
type Middleware interface {
	// Init is used to help you initialize the extension
	Init(context.Context, *graphql.Params) context.Context

	// Name returns the name of the extension (make sure it's custom)
	Name() string

	// ParseDidStart is being called before starting parsing
	ParseDidStart(context.Context) (context.Context, graphql.ParseFinishFunc)

	// ValidationDidStart is called just before validation begins
	ValidationDidStart(context.Context) (context.Context, graphql.ValidationFinishFunc)

	// ExecutionDidStart notifies about the start of the execution
	ExecutionDidStart(context.Context) (context.Context, graphql.ExecutionFinishFunc)

	// ResolveFieldDidStart notifies about the start of the resolving of a field
	ResolveFieldDidStart(context.Context, *graphql.ResolveInfo) (context.Context, graphql.ResolveFieldFinishFunc)

	// HasResult returns if the extension wants to add data to the result
	HasResult() bool

	// GetResult returns the data that the extension wants to add to the result
	GetResult(context.Context) interface{}
}
