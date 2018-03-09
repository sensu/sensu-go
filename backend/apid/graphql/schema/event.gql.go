// Code generated by scripts/gengraphql.go. DO NOT EDIT.

package schema

import (
	fmt "fmt"
	graphql1 "github.com/graphql-go/graphql"
	graphql "github.com/sensu/sensu-go/graphql"
	time "time"
)

// EventIDFieldResolver implement to resolve requests for the Event's id field.
type EventIDFieldResolver interface {
	// ID implements response to request for id field.
	ID(p graphql.ResolveParams) (interface{}, error)
}

// EventNamespaceFieldResolver implement to resolve requests for the Event's namespace field.
type EventNamespaceFieldResolver interface {
	// Namespace implements response to request for namespace field.
	Namespace(p graphql.ResolveParams) (interface{}, error)
}

// EventTimestampFieldResolver implement to resolve requests for the Event's timestamp field.
type EventTimestampFieldResolver interface {
	// Timestamp implements response to request for timestamp field.
	Timestamp(p graphql.ResolveParams) (time.Time, error)
}

// EventEntityFieldResolver implement to resolve requests for the Event's entity field.
type EventEntityFieldResolver interface {
	// Entity implements response to request for entity field.
	Entity(p graphql.ResolveParams) (interface{}, error)
}

// EventCheckFieldResolver implement to resolve requests for the Event's check field.
type EventCheckFieldResolver interface {
	// Check implements response to request for check field.
	Check(p graphql.ResolveParams) (interface{}, error)
}

// EventHooksFieldResolver implement to resolve requests for the Event's hooks field.
type EventHooksFieldResolver interface {
	// Hooks implements response to request for hooks field.
	Hooks(p graphql.ResolveParams) (interface{}, error)
}

// EventIsIncidentFieldResolver implement to resolve requests for the Event's isIncident field.
type EventIsIncidentFieldResolver interface {
	// IsIncident implements response to request for isIncident field.
	IsIncident(p graphql.ResolveParams) (bool, error)
}

// EventIsResolutionFieldResolver implement to resolve requests for the Event's isResolution field.
type EventIsResolutionFieldResolver interface {
	// IsResolution implements response to request for isResolution field.
	IsResolution(p graphql.ResolveParams) (bool, error)
}

// EventIsSilencedFieldResolver implement to resolve requests for the Event's isSilenced field.
type EventIsSilencedFieldResolver interface {
	// IsSilenced implements response to request for isSilenced field.
	IsSilenced(p graphql.ResolveParams) (bool, error)
}

//
// EventFieldResolvers represents a collection of methods whose products represent the
// response values of the 'Event' type.
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
type EventFieldResolvers interface {
	EventIDFieldResolver
	EventNamespaceFieldResolver
	EventTimestampFieldResolver
	EventEntityFieldResolver
	EventCheckFieldResolver
	EventHooksFieldResolver
	EventIsIncidentFieldResolver
	EventIsResolutionFieldResolver
	EventIsSilencedFieldResolver
}

// EventAliases implements all methods on EventFieldResolvers interface by using reflection to
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
type EventAliases struct{}

// ID implements response to request for 'id' field.
func (_ EventAliases) ID(p graphql.ResolveParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// Namespace implements response to request for 'namespace' field.
func (_ EventAliases) Namespace(p graphql.ResolveParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// Timestamp implements response to request for 'timestamp' field.
func (_ EventAliases) Timestamp(p graphql.ResolveParams) (time.Time, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret := val.(time.Time)
	return ret, err
}

// Entity implements response to request for 'entity' field.
func (_ EventAliases) Entity(p graphql.ResolveParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// Check implements response to request for 'check' field.
func (_ EventAliases) Check(p graphql.ResolveParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// Hooks implements response to request for 'hooks' field.
func (_ EventAliases) Hooks(p graphql.ResolveParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// IsIncident implements response to request for 'isIncident' field.
func (_ EventAliases) IsIncident(p graphql.ResolveParams) (bool, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret := val.(bool)
	return ret, err
}

// IsResolution implements response to request for 'isResolution' field.
func (_ EventAliases) IsResolution(p graphql.ResolveParams) (bool, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret := val.(bool)
	return ret, err
}

// IsSilenced implements response to request for 'isSilenced' field.
func (_ EventAliases) IsSilenced(p graphql.ResolveParams) (bool, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret := val.(bool)
	return ret, err
}

// EventType An Event is the encapsulating type sent across the Sensu websocket transport.
var EventType = graphql.NewType("Event", graphql.ObjectKind)

// RegisterEvent registers Event object type with given service.
func RegisterEvent(svc *graphql.Service, impl EventFieldResolvers) {
	svc.RegisterObject(_ObjectTypeEventDesc, impl)
}
func _ObjTypeEventIDHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EventIDFieldResolver)
	return func(p graphql1.ResolveParams) (interface{}, error) {
		return resolver.ID(p)
	}
}

func _ObjTypeEventNamespaceHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EventNamespaceFieldResolver)
	return func(p graphql1.ResolveParams) (interface{}, error) {
		return resolver.Namespace(p)
	}
}

func _ObjTypeEventTimestampHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EventTimestampFieldResolver)
	return func(p graphql1.ResolveParams) (interface{}, error) {
		return resolver.Timestamp(p)
	}
}

func _ObjTypeEventEntityHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EventEntityFieldResolver)
	return func(p graphql1.ResolveParams) (interface{}, error) {
		return resolver.Entity(p)
	}
}

func _ObjTypeEventCheckHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EventCheckFieldResolver)
	return func(p graphql1.ResolveParams) (interface{}, error) {
		return resolver.Check(p)
	}
}

func _ObjTypeEventHooksHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EventHooksFieldResolver)
	return func(p graphql1.ResolveParams) (interface{}, error) {
		return resolver.Hooks(p)
	}
}

func _ObjTypeEventIsIncidentHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EventIsIncidentFieldResolver)
	return func(p graphql1.ResolveParams) (interface{}, error) {
		return resolver.IsIncident(p)
	}
}

func _ObjTypeEventIsResolutionHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EventIsResolutionFieldResolver)
	return func(p graphql1.ResolveParams) (interface{}, error) {
		return resolver.IsResolution(p)
	}
}

func _ObjTypeEventIsSilencedHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EventIsSilencedFieldResolver)
	return func(p graphql1.ResolveParams) (interface{}, error) {
		return resolver.IsSilenced(p)
	}
}

func _ObjectTypeEventConfigFn() graphql1.ObjectConfig {
	return graphql1.ObjectConfig{
		Description: "An Event is the encapsulating type sent across the Sensu websocket transport.",
		Fields: graphql1.Fields{
			"check": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "Check describes the result of a check; if event is associated to check\nexecution.",
				Name:              "check",
				Type:              graphql.OutputType("Check"),
			},
			"entity": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "Entity describes the entity in which the event occurred.",
				Name:              "entity",
				Type:              graphql.OutputType("Entity"),
			},
			"hooks": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "Hooks describes the results of multiple hooks; if event is associated to hook\nexecution.",
				Name:              "hooks",
				Type:              graphql1.NewList(graphql1.NewNonNull(graphql.OutputType("Hook"))),
			},
			"id": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "The globally unique identifier of the record.",
				Name:              "id",
				Type:              graphql1.NewNonNull(graphql1.ID),
			},
			"isIncident": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "isIncident determines if an event indicates an incident.",
				Name:              "isIncident",
				Type:              graphql1.NewNonNull(graphql1.Boolean),
			},
			"isResolution": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "isResolution returns true if an event has just transitions from an incident.",
				Name:              "isResolution",
				Type:              graphql1.NewNonNull(graphql1.Boolean),
			},
			"isSilenced": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "isSilenced determines if an event has any silenced entries.",
				Name:              "isSilenced",
				Type:              graphql1.NewNonNull(graphql1.Boolean),
			},
			"namespace": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "namespace in which this record resides",
				Name:              "namespace",
				Type:              graphql1.NewNonNull(graphql.OutputType("Namespace")),
			},
			"timestamp": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "Timestamp is the time in seconds since the Epoch.",
				Name:              "timestamp",
				Type:              graphql1.NewNonNull(graphql1.DateTime),
			},
		},
		Interfaces: []*graphql1.Interface{
			graphql.Interface("Node")},
		IsTypeOf: func(_ graphql1.IsTypeOfParams) bool {
			// NOTE:
			// Panic by default. Intent is that when Service is invoked, values of
			// these fields are updated with instantiated resolvers. If these
			// defaults are called it is most certainly programmer err.
			// If you're see this comment then: 'Whoops! Sorry, my bad.'
			panic("Unimplemented; see EventFieldResolvers.")
		},
		Name: "Event",
	}
}

// describe Event's configuration; kept private to avoid unintentional tampering of configuration at runtime.
var _ObjectTypeEventDesc = graphql.ObjectDesc{
	Config: _ObjectTypeEventConfigFn,
	FieldHandlers: map[string]graphql.FieldHandler{
		"check":        _ObjTypeEventCheckHandler,
		"entity":       _ObjTypeEventEntityHandler,
		"hooks":        _ObjTypeEventHooksHandler,
		"id":           _ObjTypeEventIDHandler,
		"isIncident":   _ObjTypeEventIsIncidentHandler,
		"isResolution": _ObjTypeEventIsResolutionHandler,
		"isSilenced":   _ObjTypeEventIsSilencedHandler,
		"namespace":    _ObjTypeEventNamespaceHandler,
		"timestamp":    _ObjTypeEventTimestampHandler,
	},
}

// EventConnectionEdgesFieldResolver implement to resolve requests for the EventConnection's edges field.
type EventConnectionEdgesFieldResolver interface {
	// Edges implements response to request for edges field.
	Edges(p graphql.ResolveParams) (interface{}, error)
}

// EventConnectionPageInfoFieldResolver implement to resolve requests for the EventConnection's pageInfo field.
type EventConnectionPageInfoFieldResolver interface {
	// PageInfo implements response to request for pageInfo field.
	PageInfo(p graphql.ResolveParams) (interface{}, error)
}

// EventConnectionTotalCountFieldResolver implement to resolve requests for the EventConnection's totalCount field.
type EventConnectionTotalCountFieldResolver interface {
	// TotalCount implements response to request for totalCount field.
	TotalCount(p graphql.ResolveParams) (int, error)
}

//
// EventConnectionFieldResolvers represents a collection of methods whose products represent the
// response values of the 'EventConnection' type.
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
type EventConnectionFieldResolvers interface {
	EventConnectionEdgesFieldResolver
	EventConnectionPageInfoFieldResolver
	EventConnectionTotalCountFieldResolver
}

// EventConnectionAliases implements all methods on EventConnectionFieldResolvers interface by using reflection to
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
type EventConnectionAliases struct{}

// Edges implements response to request for 'edges' field.
func (_ EventConnectionAliases) Edges(p graphql.ResolveParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// PageInfo implements response to request for 'pageInfo' field.
func (_ EventConnectionAliases) PageInfo(p graphql.ResolveParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// TotalCount implements response to request for 'totalCount' field.
func (_ EventConnectionAliases) TotalCount(p graphql.ResolveParams) (int, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret := graphql1.Int.ParseValue(val).(int)
	return ret, err
}

// EventConnectionType A connection to a sequence of records.
var EventConnectionType = graphql.NewType("EventConnection", graphql.ObjectKind)

// RegisterEventConnection registers EventConnection object type with given service.
func RegisterEventConnection(svc *graphql.Service, impl EventConnectionFieldResolvers) {
	svc.RegisterObject(_ObjectTypeEventConnectionDesc, impl)
}
func _ObjTypeEventConnectionEdgesHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EventConnectionEdgesFieldResolver)
	return func(p graphql1.ResolveParams) (interface{}, error) {
		return resolver.Edges(p)
	}
}

func _ObjTypeEventConnectionPageInfoHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EventConnectionPageInfoFieldResolver)
	return func(p graphql1.ResolveParams) (interface{}, error) {
		return resolver.PageInfo(p)
	}
}

func _ObjTypeEventConnectionTotalCountHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EventConnectionTotalCountFieldResolver)
	return func(p graphql1.ResolveParams) (interface{}, error) {
		return resolver.TotalCount(p)
	}
}

func _ObjectTypeEventConnectionConfigFn() graphql1.ObjectConfig {
	return graphql1.ObjectConfig{
		Description: "A connection to a sequence of records.",
		Fields: graphql1.Fields{
			"edges": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "self descriptive",
				Name:              "edges",
				Type:              graphql1.NewList(graphql.OutputType("EventEdge")),
			},
			"pageInfo": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "self descriptive",
				Name:              "pageInfo",
				Type:              graphql1.NewNonNull(graphql.OutputType("PageInfo")),
			},
			"totalCount": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "self descriptive",
				Name:              "totalCount",
				Type:              graphql1.NewNonNull(graphql1.Int),
			},
		},
		Interfaces: []*graphql1.Interface{},
		IsTypeOf: func(_ graphql1.IsTypeOfParams) bool {
			// NOTE:
			// Panic by default. Intent is that when Service is invoked, values of
			// these fields are updated with instantiated resolvers. If these
			// defaults are called it is most certainly programmer err.
			// If you're see this comment then: 'Whoops! Sorry, my bad.'
			panic("Unimplemented; see EventConnectionFieldResolvers.")
		},
		Name: "EventConnection",
	}
}

// describe EventConnection's configuration; kept private to avoid unintentional tampering of configuration at runtime.
var _ObjectTypeEventConnectionDesc = graphql.ObjectDesc{
	Config: _ObjectTypeEventConnectionConfigFn,
	FieldHandlers: map[string]graphql.FieldHandler{
		"edges":      _ObjTypeEventConnectionEdgesHandler,
		"pageInfo":   _ObjTypeEventConnectionPageInfoHandler,
		"totalCount": _ObjTypeEventConnectionTotalCountHandler,
	},
}

// EventEdgeNodeFieldResolver implement to resolve requests for the EventEdge's node field.
type EventEdgeNodeFieldResolver interface {
	// Node implements response to request for node field.
	Node(p graphql.ResolveParams) (interface{}, error)
}

// EventEdgeCursorFieldResolver implement to resolve requests for the EventEdge's cursor field.
type EventEdgeCursorFieldResolver interface {
	// Cursor implements response to request for cursor field.
	Cursor(p graphql.ResolveParams) (string, error)
}

//
// EventEdgeFieldResolvers represents a collection of methods whose products represent the
// response values of the 'EventEdge' type.
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
type EventEdgeFieldResolvers interface {
	EventEdgeNodeFieldResolver
	EventEdgeCursorFieldResolver
}

// EventEdgeAliases implements all methods on EventEdgeFieldResolvers interface by using reflection to
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
type EventEdgeAliases struct{}

// Node implements response to request for 'node' field.
func (_ EventEdgeAliases) Node(p graphql.ResolveParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// Cursor implements response to request for 'cursor' field.
func (_ EventEdgeAliases) Cursor(p graphql.ResolveParams) (string, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret := fmt.Sprint(val)
	return ret, err
}

// EventEdgeType An edge in a connection.
var EventEdgeType = graphql.NewType("EventEdge", graphql.ObjectKind)

// RegisterEventEdge registers EventEdge object type with given service.
func RegisterEventEdge(svc *graphql.Service, impl EventEdgeFieldResolvers) {
	svc.RegisterObject(_ObjectTypeEventEdgeDesc, impl)
}
func _ObjTypeEventEdgeNodeHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EventEdgeNodeFieldResolver)
	return func(p graphql1.ResolveParams) (interface{}, error) {
		return resolver.Node(p)
	}
}

func _ObjTypeEventEdgeCursorHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EventEdgeCursorFieldResolver)
	return func(p graphql1.ResolveParams) (interface{}, error) {
		return resolver.Cursor(p)
	}
}

func _ObjectTypeEventEdgeConfigFn() graphql1.ObjectConfig {
	return graphql1.ObjectConfig{
		Description: "An edge in a connection.",
		Fields: graphql1.Fields{
			"cursor": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "self descriptive",
				Name:              "cursor",
				Type:              graphql1.NewNonNull(graphql1.String),
			},
			"node": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "self descriptive",
				Name:              "node",
				Type:              graphql.OutputType("Event"),
			},
		},
		Interfaces: []*graphql1.Interface{},
		IsTypeOf: func(_ graphql1.IsTypeOfParams) bool {
			// NOTE:
			// Panic by default. Intent is that when Service is invoked, values of
			// these fields are updated with instantiated resolvers. If these
			// defaults are called it is most certainly programmer err.
			// If you're see this comment then: 'Whoops! Sorry, my bad.'
			panic("Unimplemented; see EventEdgeFieldResolvers.")
		},
		Name: "EventEdge",
	}
}

// describe EventEdge's configuration; kept private to avoid unintentional tampering of configuration at runtime.
var _ObjectTypeEventEdgeDesc = graphql.ObjectDesc{
	Config: _ObjectTypeEventEdgeConfigFn,
	FieldHandlers: map[string]graphql.FieldHandler{
		"cursor": _ObjTypeEventEdgeCursorHandler,
		"node":   _ObjTypeEventEdgeNodeHandler,
	},
}
