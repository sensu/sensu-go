package graphql

import (
	"context"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
)

// Middleware is an interface for extending a service with operations that
// wrap existing functionality.
type Middleware interface {
	// Init is called at the beginning of an execution. Helpful for initializing
	// anything used during the execution of the query.
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

type (
	// parseFinishFuncHandler handles the call of all the ParseFinishFuncs from the extenisons
	parseFinishFuncHandler func(error)

	// validationFinishFuncHandler responsible for the call of all the ValidationFinishFuncs
	validationFinishFuncHandler func([]gqlerrors.FormattedError)

	// executionFinishFuncHandler calls all the ExecutionFinishFuncs from each extension
	executionFinishFuncHandler func(*graphql.Result)

	// resolveFieldFinishFuncHandler calls the resolveFieldFinishFns for all the extensions
	resolveFieldFinishFuncHandler func(interface{}, error)
)

// MiddlewareHandleInits handles all the init functions for the given service.
func MiddlewareHandleInits(s *Service, p *graphql.Params) {
	for _, m := range s.mware {
		// update context
		p.Context = m.Init(p.Context, p)
	}
}

// MiddlewareHandleParseDidStart runs the ParseDidStart functions for each extension
func MiddlewareHandleParseDidStart(s *Service, p *graphql.Params) parseFinishFuncHandler {
	fs := map[string]graphql.ParseFinishFunc{}
	for _, m := range s.mware {
		ctx, finishFn := m.ParseDidStart(p.Context)
		p.Context = ctx
		fs[m.Name()] = finishFn
	}
	return func(err error) {
		for _, fn := range fs {
			fn(err)
		}
	}
}

// MiddlewareHandleValidationDidStart notifies the extensions about the start of the validation process
func MiddlewareHandleValidationDidStart(s *Service, p *graphql.Params) validationFinishFuncHandler {
	fs := map[string]graphql.ValidationFinishFunc{}
	for _, m := range s.mware {
		ctx, finishFn := m.ValidationDidStart(p.Context)
		p.Context = ctx
		fs[m.Name()] = finishFn
	}
	return func(errs []gqlerrors.FormattedError) {
		for _, finishFn := range fs {
			finishFn(errs)
		}
	}
}

// MiddlewareHandleExecutionDidStart handles the ExecutionDidStart func
func MiddlewareHandleExecutionDidStart(s *Service, p *graphql.ExecuteParams) executionFinishFuncHandler {
	fs := map[string]graphql.ExecutionFinishFunc{}
	for _, m := range s.mware {
		ctx, finishFn := m.ExecutionDidStart(p.Context)
		p.Context = ctx
		fs[m.Name()] = finishFn
	}
	return func(result *graphql.Result) {
		for _, finishFn := range fs {
			finishFn(result)
		}
	}
}

// MiddlewareHandleResolveFieldDidStart handles the notification of the extensions about the start of a resolve function
func MiddlewareHandleResolveFieldDidStart(ctx context.Context, mware []Middleware, i *ResolveInfo) (context.Context, resolveFieldFinishFuncHandler) {
	fs := map[string]graphql.ResolveFieldFinishFunc{}
	for _, m := range mware {
		var finishFn graphql.ResolveFieldFinishFunc
		ctx, finishFn = m.ResolveFieldDidStart(ctx, i)
		fs[m.Name()] = finishFn
	}
	return ctx, func(val interface{}, err error) {
		for _, finishFn := range fs {
			finishFn(val, err)
		}
	}
}
