package graphql

import (
	"github.com/graphql-go/graphql"
)

// ResolveTypeParams used in ResolveType fn
type ResolveTypeParams = graphql.ResolveTypeParams

// IsTypeOfParams used in IsTypeOf fn
type IsTypeOfParams = graphql.IsTypeOfParams

// ResolveParams params for field resolvers
type ResolveParams = graphql.ResolveParams

// ResolveInfo is a collection of information about the current execution state
type ResolveInfo = graphql.ResolveInfo

// FieldHandler given implementation configures field resolver
type FieldHandler func(impl interface{}) graphql.FieldResolveFn

// Result has the response, errors and extensions from the resolved schema
type Result = graphql.Result

// ObjectDesc describes object configuration and handlers for use by service.
type ObjectDesc struct {
	// Config thunk returns copy of config
	Config func() graphql.ObjectConfig
	// FieldHandlers handlers that wrap each field resolver.
	FieldHandlers map[string]FieldHandler
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
	Config func() graphql.InputObjectConfig
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
