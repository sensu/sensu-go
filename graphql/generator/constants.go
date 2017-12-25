package generator

const (
	// ast package
	astPkg = "github.com/graphql-go/graphql/language/ast"

	// graphql-go package; used in generated code for definition types
	defsPkg = "github.com/graphql-go/graphql"

	// GraphQL service package
	servicePkg = "github.com/sensu/sensu-go/graphql"

	// used to describe resolverFns that panic when not implemented.
	missingResolverNote = `// NOTE:
// Panic by default. Intent is that when Service is invoked, values of
// these fields are updated with instantiated resolvers. If these
// defaults are called it is most certainly programmer err.
// If you're see this comment then: 'Whoops! Sorry, my bad.'`
)
