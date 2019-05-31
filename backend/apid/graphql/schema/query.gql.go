// Code generated by scripts/gengraphql.go. DO NOT EDIT.

package schema

import (
	graphql1 "github.com/graphql-go/graphql"
	mapstructure "github.com/mitchellh/mapstructure"
	graphql "github.com/sensu/sensu-go/graphql"
)

// QueryViewerFieldResolver implement to resolve requests for the Query's viewer field.
type QueryViewerFieldResolver interface {
	// Viewer implements response to request for viewer field.
	Viewer(p graphql.ResolveParams) (interface{}, error)
}

// QueryNamespaceFieldResolverArgs contains arguments provided to namespace when selected
type QueryNamespaceFieldResolverArgs struct {
	Name string // Name - self descriptive
}

// QueryNamespaceFieldResolverParams contains contextual info to resolve namespace field
type QueryNamespaceFieldResolverParams struct {
	graphql.ResolveParams
	Args QueryNamespaceFieldResolverArgs
}

// QueryNamespaceFieldResolver implement to resolve requests for the Query's namespace field.
type QueryNamespaceFieldResolver interface {
	// Namespace implements response to request for namespace field.
	Namespace(p QueryNamespaceFieldResolverParams) (interface{}, error)
}

// QueryEventFieldResolverArgs contains arguments provided to event when selected
type QueryEventFieldResolverArgs struct {
	Namespace string // Namespace - self descriptive
	Entity    string // Entity - self descriptive
	Check     string // Check - self descriptive
}

// QueryEventFieldResolverParams contains contextual info to resolve event field
type QueryEventFieldResolverParams struct {
	graphql.ResolveParams
	Args QueryEventFieldResolverArgs
}

// QueryEventFieldResolver implement to resolve requests for the Query's event field.
type QueryEventFieldResolver interface {
	// Event implements response to request for event field.
	Event(p QueryEventFieldResolverParams) (interface{}, error)
}

// QueryEntityFieldResolverArgs contains arguments provided to entity when selected
type QueryEntityFieldResolverArgs struct {
	Namespace string // Namespace - self descriptive
	Name      string // Name - self descriptive
}

// QueryEntityFieldResolverParams contains contextual info to resolve entity field
type QueryEntityFieldResolverParams struct {
	graphql.ResolveParams
	Args QueryEntityFieldResolverArgs
}

// QueryEntityFieldResolver implement to resolve requests for the Query's entity field.
type QueryEntityFieldResolver interface {
	// Entity implements response to request for entity field.
	Entity(p QueryEntityFieldResolverParams) (interface{}, error)
}

// QueryCheckFieldResolverArgs contains arguments provided to check when selected
type QueryCheckFieldResolverArgs struct {
	Namespace string // Namespace - self descriptive
	Name      string // Name - self descriptive
}

// QueryCheckFieldResolverParams contains contextual info to resolve check field
type QueryCheckFieldResolverParams struct {
	graphql.ResolveParams
	Args QueryCheckFieldResolverArgs
}

// QueryCheckFieldResolver implement to resolve requests for the Query's check field.
type QueryCheckFieldResolver interface {
	// Check implements response to request for check field.
	Check(p QueryCheckFieldResolverParams) (interface{}, error)
}

// QueryHandlerFieldResolverArgs contains arguments provided to handler when selected
type QueryHandlerFieldResolverArgs struct {
	Namespace string // Namespace - self descriptive
	Name      string // Name - self descriptive
}

// QueryHandlerFieldResolverParams contains contextual info to resolve handler field
type QueryHandlerFieldResolverParams struct {
	graphql.ResolveParams
	Args QueryHandlerFieldResolverArgs
}

// QueryHandlerFieldResolver implement to resolve requests for the Query's handler field.
type QueryHandlerFieldResolver interface {
	// Handler implements response to request for handler field.
	Handler(p QueryHandlerFieldResolverParams) (interface{}, error)
}

// QuerySuggestsFieldResolverArgs contains arguments provided to suggests when selected
type QuerySuggestsFieldResolverArgs struct {
	Namespace string          // Namespace - self descriptive
	Ref       string          // Ref - self descriptive
	Includes  string          // Includes - self descriptive
	Limit     int             // Limit - self descriptive
	Order     SuggestionOrder // Order - self descriptive
}

// QuerySuggestsFieldResolverParams contains contextual info to resolve suggests field
type QuerySuggestsFieldResolverParams struct {
	graphql.ResolveParams
	Args QuerySuggestsFieldResolverArgs
}

// QuerySuggestsFieldResolver implement to resolve requests for the Query's suggests field.
type QuerySuggestsFieldResolver interface {
	// Suggests implements response to request for suggests field.
	Suggests(p QuerySuggestsFieldResolverParams) (interface{}, error)
}

// QueryNodeFieldResolverArgs contains arguments provided to node when selected
type QueryNodeFieldResolverArgs struct {
	ID string // ID - The ID of an object.
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

// QueryWrappedNodeFieldResolverArgs contains arguments provided to wrappedNode when selected
type QueryWrappedNodeFieldResolverArgs struct {
	ID string // ID - The ID of an object.
}

// QueryWrappedNodeFieldResolverParams contains contextual info to resolve wrappedNode field
type QueryWrappedNodeFieldResolverParams struct {
	graphql.ResolveParams
	Args QueryWrappedNodeFieldResolverArgs
}

// QueryWrappedNodeFieldResolver implement to resolve requests for the Query's wrappedNode field.
type QueryWrappedNodeFieldResolver interface {
	// WrappedNode implements response to request for wrappedNode field.
	WrappedNode(p QueryWrappedNodeFieldResolverParams) (interface{}, error)
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
	QueryNamespaceFieldResolver
	QueryEventFieldResolver
	QueryEntityFieldResolver
	QueryCheckFieldResolver
	QueryHandlerFieldResolver
	QuerySuggestsFieldResolver
	QueryNodeFieldResolver
	QueryWrappedNodeFieldResolver
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

// Namespace implements response to request for 'namespace' field.
func (_ QueryAliases) Namespace(p QueryNamespaceFieldResolverParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// Event implements response to request for 'event' field.
func (_ QueryAliases) Event(p QueryEventFieldResolverParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// Entity implements response to request for 'entity' field.
func (_ QueryAliases) Entity(p QueryEntityFieldResolverParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// Check implements response to request for 'check' field.
func (_ QueryAliases) Check(p QueryCheckFieldResolverParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// Handler implements response to request for 'handler' field.
func (_ QueryAliases) Handler(p QueryHandlerFieldResolverParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// Suggests implements response to request for 'suggests' field.
func (_ QueryAliases) Suggests(p QuerySuggestsFieldResolverParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// Node implements response to request for 'node' field.
func (_ QueryAliases) Node(p QueryNodeFieldResolverParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// WrappedNode implements response to request for 'wrappedNode' field.
func (_ QueryAliases) WrappedNode(p QueryWrappedNodeFieldResolverParams) (interface{}, error) {
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
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Viewer(frp)
	}
}

func _ObjTypeQueryNamespaceHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(QueryNamespaceFieldResolver)
	return func(p graphql1.ResolveParams) (interface{}, error) {
		frp := QueryNamespaceFieldResolverParams{ResolveParams: p}
		err := mapstructure.Decode(p.Args, &frp.Args)
		if err != nil {
			return nil, err
		}

		return resolver.Namespace(frp)
	}
}

func _ObjTypeQueryEventHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(QueryEventFieldResolver)
	return func(p graphql1.ResolveParams) (interface{}, error) {
		frp := QueryEventFieldResolverParams{ResolveParams: p}
		err := mapstructure.Decode(p.Args, &frp.Args)
		if err != nil {
			return nil, err
		}

		return resolver.Event(frp)
	}
}

func _ObjTypeQueryEntityHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(QueryEntityFieldResolver)
	return func(p graphql1.ResolveParams) (interface{}, error) {
		frp := QueryEntityFieldResolverParams{ResolveParams: p}
		err := mapstructure.Decode(p.Args, &frp.Args)
		if err != nil {
			return nil, err
		}

		return resolver.Entity(frp)
	}
}

func _ObjTypeQueryCheckHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(QueryCheckFieldResolver)
	return func(p graphql1.ResolveParams) (interface{}, error) {
		frp := QueryCheckFieldResolverParams{ResolveParams: p}
		err := mapstructure.Decode(p.Args, &frp.Args)
		if err != nil {
			return nil, err
		}

		return resolver.Check(frp)
	}
}

func _ObjTypeQueryHandlerHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(QueryHandlerFieldResolver)
	return func(p graphql1.ResolveParams) (interface{}, error) {
		frp := QueryHandlerFieldResolverParams{ResolveParams: p}
		err := mapstructure.Decode(p.Args, &frp.Args)
		if err != nil {
			return nil, err
		}

		return resolver.Handler(frp)
	}
}

func _ObjTypeQuerySuggestsHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(QuerySuggestsFieldResolver)
	return func(p graphql1.ResolveParams) (interface{}, error) {
		frp := QuerySuggestsFieldResolverParams{ResolveParams: p}
		err := mapstructure.Decode(p.Args, &frp.Args)
		if err != nil {
			return nil, err
		}

		return resolver.Suggests(frp)
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

func _ObjTypeQueryWrappedNodeHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(QueryWrappedNodeFieldResolver)
	return func(p graphql1.ResolveParams) (interface{}, error) {
		frp := QueryWrappedNodeFieldResolverParams{ResolveParams: p}
		err := mapstructure.Decode(p.Args, &frp.Args)
		if err != nil {
			return nil, err
		}

		return resolver.WrappedNode(frp)
	}
}

func _ObjectTypeQueryConfigFn() graphql1.ObjectConfig {
	return graphql1.ObjectConfig{
		Description: "The query root of Sensu's GraphQL interface.",
		Fields: graphql1.Fields{
			"check": &graphql1.Field{
				Args: graphql1.FieldConfigArgument{
					"name": &graphql1.ArgumentConfig{
						Description: "self descriptive",
						Type:        graphql1.NewNonNull(graphql1.String),
					},
					"namespace": &graphql1.ArgumentConfig{
						Description: "self descriptive",
						Type:        graphql1.NewNonNull(graphql1.String),
					},
				},
				DeprecationReason: "",
				Description:       "check fetches the check config associated with the given set of arguments.",
				Name:              "check",
				Type:              graphql.OutputType("CheckConfig"),
			},
			"entity": &graphql1.Field{
				Args: graphql1.FieldConfigArgument{
					"name": &graphql1.ArgumentConfig{
						Description: "self descriptive",
						Type:        graphql1.NewNonNull(graphql1.String),
					},
					"namespace": &graphql1.ArgumentConfig{
						Description: "self descriptive",
						Type:        graphql1.NewNonNull(graphql1.String),
					},
				},
				DeprecationReason: "",
				Description:       "Entity fetches the entity associated with the given set of arguments.",
				Name:              "entity",
				Type:              graphql.OutputType("Entity"),
			},
			"event": &graphql1.Field{
				Args: graphql1.FieldConfigArgument{
					"check": &graphql1.ArgumentConfig{
						Description: "self descriptive",
						Type:        graphql1.String,
					},
					"entity": &graphql1.ArgumentConfig{
						Description: "self descriptive",
						Type:        graphql1.NewNonNull(graphql1.String),
					},
					"namespace": &graphql1.ArgumentConfig{
						Description: "self descriptive",
						Type:        graphql1.NewNonNull(graphql1.String),
					},
				},
				DeprecationReason: "",
				Description:       "Event fetches the event associated with the given set of arguments.",
				Name:              "event",
				Type:              graphql.OutputType("Event"),
			},
			"handler": &graphql1.Field{
				Args: graphql1.FieldConfigArgument{
					"name": &graphql1.ArgumentConfig{
						Description: "self descriptive",
						Type:        graphql1.NewNonNull(graphql1.String),
					},
					"namespace": &graphql1.ArgumentConfig{
						Description: "self descriptive",
						Type:        graphql1.NewNonNull(graphql1.String),
					},
				},
				DeprecationReason: "",
				Description:       "handler fetch the handler associated with the given set of arguments.",
				Name:              "handler",
				Type:              graphql.OutputType("Handler"),
			},
			"namespace": &graphql1.Field{
				Args: graphql1.FieldConfigArgument{"name": &graphql1.ArgumentConfig{
					Description: "self descriptive",
					Type:        graphql1.NewNonNull(graphql1.String),
				}},
				DeprecationReason: "",
				Description:       "Namespace fetches the namespace object associated with the given name.",
				Name:              "namespace",
				Type:              graphql.OutputType("Namespace"),
			},
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
			"suggests": &graphql1.Field{
				Args: graphql1.FieldConfigArgument{
					"includes": &graphql1.ArgumentConfig{
						DefaultValue: "",
						Description:  "self descriptive",
						Type:         graphql1.String,
					},
					"limit": &graphql1.ArgumentConfig{
						DefaultValue: 10,
						Description:  "self descriptive",
						Type:         graphql1.Int,
					},
					"namespace": &graphql1.ArgumentConfig{
						Description: "self descriptive",
						Type:        graphql1.NewNonNull(graphql1.String),
					},
					"order": &graphql1.ArgumentConfig{
						DefaultValue: "FREQUENCY",
						Description:  "self descriptive",
						Type:         graphql.InputType("SuggestionOrder"),
					},
					"ref": &graphql1.ArgumentConfig{
						Description: "self descriptive",
						Type:        graphql1.NewNonNull(graphql1.String),
					},
				},
				DeprecationReason: "",
				Description:       "Given a type, field and a namespace returns a set of suggested values.",
				Name:              "suggests",
				Type:              graphql.OutputType("SuggestionResultSet"),
			},
			"viewer": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "Current viewer.",
				Name:              "viewer",
				Type:              graphql.OutputType("Viewer"),
			},
			"wrappedNode": &graphql1.Field{
				Args: graphql1.FieldConfigArgument{"id": &graphql1.ArgumentConfig{
					Description: "The ID of an object.",
					Type:        graphql1.NewNonNull(graphql1.ID),
				}},
				DeprecationReason: "",
				Description:       "Node fetches an object given its ID and returns it as wrapped resource.",
				Name:              "wrappedNode",
				Type:              graphql.OutputType("JSON"),
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
		"check":       _ObjTypeQueryCheckHandler,
		"entity":      _ObjTypeQueryEntityHandler,
		"event":       _ObjTypeQueryEventHandler,
		"handler":     _ObjTypeQueryHandlerHandler,
		"namespace":   _ObjTypeQueryNamespaceHandler,
		"node":        _ObjTypeQueryNodeHandler,
		"suggests":    _ObjTypeQuerySuggestsHandler,
		"viewer":      _ObjTypeQueryViewerHandler,
		"wrappedNode": _ObjTypeQueryWrappedNodeHandler,
	},
}
