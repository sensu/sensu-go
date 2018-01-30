// Code generated by scripts/gengraphql.go. DO NOT EDIT.

package schema

import (
	graphql1 "github.com/graphql-go/graphql"
	mapstructure "github.com/mitchellh/mapstructure"
	graphql "github.com/sensu/sensu-go/graphql"
)

// Schema supplies the root types of each type of operation, query,
// mutation (optional), and subscription (optional).
var Schema = graphql.NewType("Schema", graphql.SchemaKind)

// RegisterSchema registers schema description with given service.
func RegisterSchema(svc *graphql.Service) {
	svc.RegisterSchema(_SchemaDesc)
}
func _SchemaConfigFn() graphql1.SchemaConfig {
	return graphql1.SchemaConfig{
		Mutation: graphql.Object("Mutation"),
		Query:    graphql.Object("Query"),
	}
}

// describe schema's configuration; kept private to avoid unintentional tampering of configuration at runtime.
var _SchemaDesc = graphql.SchemaDesc{Config: _SchemaConfigFn}

// QueryViewerFieldResolver implement to resolve requests for the Query's viewer field.
type QueryViewerFieldResolver interface {
	// Viewer implements response to request for viewer field.
	Viewer(p graphql.ResolveParams) (interface{}, error)
}

// QueryNodeFieldResolverArgs contains arguments provided to node when selected
type QueryNodeFieldResolverArgs struct {
	ID interface{} // ID - The ID of an object.
}

// QueryNodeFieldResolverParams contains contextual info to resolve node field
type QueryNodeFieldResolverParams struct {
	graphql.ResolveParams
	Args QueryNodeFieldResolverArgs
}

// QueryNodeFieldResolver implement to resolve requests for the Query's node field.
type QueryNodeFieldResolver interface {
	// Node implements response to request for node field.
	Node(p QueryNodeFieldResolverParams) (interface{}, error)
}

//
// QueryFieldResolvers represents a collection of methods whose products represent the
// response values of the 'Query' type.
//
// == Example SDL
//
//   """
//   Dog's are not hooman.
//   """
//   type Dog implements Pet {
//     "name of this fine beast."
//     name:  String!
//
//     "breed of this silly animal; probably shibe."
//     breed: [Breed]
//   }
//
// == Example generated interface
//
//   // DogResolver ...
//   type DogFieldResolvers interface {
//     DogNameFieldResolver
//     DogBreedFieldResolver
//
//     // IsTypeOf is used to determine if a given value is associated with the Dog type
//     IsTypeOf(interface{}, graphql.IsTypeOfParams) bool
//   }
//
// == Example implementation ...
//
//   // DogResolver implements DogFieldResolvers interface
//   type DogResolver struct {
//     logger logrus.LogEntry
//     store interface{
//       store.BreedStore
//       store.DogStore
//     }
//   }
//
//   // Name implements response to request for name field.
//   func (r *DogResolver) Name(p graphql.ResolveParams) (interface{}, error) {
//     // ... implementation details ...
//     dog := p.Source.(DogGetter)
//     return dog.GetName()
//   }
//
//   // Breed implements response to request for breed field.
//   func (r *DogResolver) Breed(p graphql.ResolveParams) (interface{}, error) {
//     // ... implementation details ...
//     dog := p.Source.(DogGetter)
//     breed := r.store.GetBreed(dog.GetBreedName())
//     return breed
//   }
//
//   // IsTypeOf is used to determine if a given value is associated with the Dog type
//   func (r *DogResolver) IsTypeOf(p graphql.IsTypeOfParams) bool {
//     // ... implementation details ...
//     _, ok := p.Value.(DogGetter)
//     return ok
//   }
//
type QueryFieldResolvers interface {
	QueryViewerFieldResolver
	QueryNodeFieldResolver
}

// QueryAliases implements all methods on QueryFieldResolvers interface by using reflection to
// match name of field to a field on the given value. Intent is reduce friction
// of writing new resolvers by removing all the instances where you would simply
// have the resolvers method return a field.
//
// == Example SDL
//
//    type Dog {
//      name:   String!
//      weight: Float!
//      dob:    DateTime
//      breed:  [Breed]
//    }
//
// == Example generated aliases
//
//   type DogAliases struct {}
//   func (_ DogAliases) Name(p graphql.ResolveParams) (interface{}, error) {
//     // reflect...
//   }
//   func (_ DogAliases) Weight(p graphql.ResolveParams) (interface{}, error) {
//     // reflect...
//   }
//   func (_ DogAliases) Dob(p graphql.ResolveParams) (interface{}, error) {
//     // reflect...
//   }
//   func (_ DogAliases) Breed(p graphql.ResolveParams) (interface{}, error) {
//     // reflect...
//   }
//
// == Example Implementation
//
//   type DogResolver struct { // Implements DogResolver
//     DogAliases
//     store store.BreedStore
//   }
//
//   // NOTE:
//   // All other fields are satisified by DogAliases but since this one
//   // requires hitting the store we implement it in our resolver.
//   func (r *DogResolver) Breed(p graphql.ResolveParams) interface{} {
//     dog := v.(*Dog)
//     return r.BreedsById(dog.BreedIDs)
//   }
//
type QueryAliases struct{}

// Viewer implements response to request for 'viewer' field.
func (_ QueryAliases) Viewer(p graphql.ResolveParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// Node implements response to request for 'node' field.
func (_ QueryAliases) Node(p QueryNodeFieldResolverParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// QueryType The query root of Sensu's GraphQL interface.
var QueryType = graphql.NewType("Query", graphql.ObjectKind)

// RegisterQuery registers Query object type with given service.
func RegisterQuery(svc *graphql.Service, impl QueryFieldResolvers) {
	svc.RegisterObject(_ObjectTypeQueryDesc, impl)
}
func _ObjTypeQueryViewerHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(QueryViewerFieldResolver)
	return func(p graphql1.ResolveParams) (interface{}, error) {
		return resolver.Viewer(p)
	}
}

func _ObjTypeQueryNodeHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(QueryNodeFieldResolver)
	return func(p graphql1.ResolveParams) (interface{}, error) {
		frp := QueryNodeFieldResolverParams{ResolveParams: p}
		err := mapstructure.Decode(p.Args, &frp.Args)
		if err != nil {
			return nil, err
		}

		return resolver.Node(frp)
	}
}

func _ObjectTypeQueryConfigFn() graphql1.ObjectConfig {
	return graphql1.ObjectConfig{
		Description: "The query root of Sensu's GraphQL interface.",
		Fields: graphql1.Fields{
			"node": &graphql1.Field{
				Args: graphql1.FieldConfigArgument{"id": &graphql1.ArgumentConfig{
					Description: "The ID of an object.",
					Type:        graphql1.NewNonNull(graphql1.ID),
				}},
				DeprecationReason: "",
				Description:       "Node fetches an object given its ID.",
				Name:              "node",
				Type:              graphql.OutputType("Node"),
			},
			"viewer": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "Current viewer.",
				Name:              "viewer",
				Type:              graphql.OutputType("Viewer"),
			},
		},
		Interfaces: []*graphql1.Interface{},
		IsTypeOf: func(_ graphql1.IsTypeOfParams) bool {
			// NOTE:
			// Panic by default. Intent is that when Service is invoked, values of
			// these fields are updated with instantiated resolvers. If these
			// defaults are called it is most certainly programmer err.
			// If you're see this comment then: 'Whoops! Sorry, my bad.'
			panic("Unimplemented; see QueryFieldResolvers.")
		},
		Name: "Query",
	}
}

// describe Query's configuration; kept private to avoid unintentional tampering of configuration at runtime.
var _ObjectTypeQueryDesc = graphql.ObjectDesc{
	Config: _ObjectTypeQueryConfigFn,
	FieldHandlers: map[string]graphql.FieldHandler{
		"node":   _ObjTypeQueryNodeHandler,
		"viewer": _ObjTypeQueryViewerHandler,
	},
}
