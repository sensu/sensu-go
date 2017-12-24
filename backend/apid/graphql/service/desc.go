package service

import "github.com/graphql-go/graphql"

// ResolveTypeParams used in ResolveType fn
type ResolveTypeParams = graphql.ResolveTypeParams

// IsTypeOfParams used in IsTypeOf fn
type IsTypeOfParams = graphql.IsTypeOfParams

// ResolveContext describes contextual information about current query / field
type ResolveContext struct {
	// Info is a collection of information about the current execution state.
	Info graphql.ResolveInfo

	// Context argument is a context value that is provided to every resolve function within an execution.
	// It is commonly
	// used to represent an authenticated user, or request-specific caches.
	Context context.Context
}

// ResolveParams params for field resolvers
type ResolveParams struct {
	ResolveContext

	// Source is the source value
	Source interface{}
}

// ObjectDesc describes object configuration and handlers for use by service.
type ObjectDesc struct {
	// Config thunk returns copy of config
	Config func() graphql.ObjectConfig
	// Collection of handlers for each field
	ResolverHandlers map[string]func(interface{}, graphql.ResolveParams)
}

// ScalarDesc describes scalar configuration and handlers for use by service.
type ScalarDesc struct {
	// Config thunk returns copy of config
	Config func() graphql.ScalarConfig
}

// UnionDesc describes union configuration and handlers for use by service.
type UnionDesc struct {
	// Config thunk returns copy of config
	Config func() graphql.UnionConfig
}

// EnumDesc describes enum configuration and handlers for use by service.
type EnumDesc struct {
	// Config thunk returns copy of config
	Config func() graphql.EnumConfig
}

// InputDesc describes input configuration and handlers for use by service.
type InputDesc struct {
	// Config thunk returns copy of config
	Config func() graphql.InputConfig
}

// InterfaceDesc describes interface configuration and handlers for use by service.
type InterfaceDesc struct {
	// Config thunk returns copy of config
	Config func() graphql.InterfaceConfig
}

// SchemaDesc describes schema configuration and handlers for use by service.
type SchemaDesc struct {
	// Config thunk returns copy of config
	Config func() graphql.SchemaConfig
}
