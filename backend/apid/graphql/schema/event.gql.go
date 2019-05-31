// Code generated by scripts/gengraphql.go. DO NOT EDIT.

package schema

import (
	errors "errors"
	graphql1 "github.com/graphql-go/graphql"
	graphql "github.com/sensu/sensu-go/graphql"
	time "time"
)

// EventIDFieldResolver implement to resolve requests for the Event's id field.
type EventIDFieldResolver interface {
	// ID implements response to request for id field.
	ID(p graphql.ResolveParams) (string, error)
}

// EventNamespaceFieldResolver implement to resolve requests for the Event's namespace field.
type EventNamespaceFieldResolver interface {
	// Namespace implements response to request for namespace field.
	Namespace(p graphql.ResolveParams) (string, error)
}

// EventMetadataFieldResolver implement to resolve requests for the Event's metadata field.
type EventMetadataFieldResolver interface {
	// Metadata implements response to request for metadata field.
	Metadata(p graphql.ResolveParams) (interface{}, error)
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

// EventIsNewIncidentFieldResolver implement to resolve requests for the Event's isNewIncident field.
type EventIsNewIncidentFieldResolver interface {
	// IsNewIncident implements response to request for isNewIncident field.
	IsNewIncident(p graphql.ResolveParams) (bool, error)
}

// EventIsResolutionFieldResolver implement to resolve requests for the Event's isResolution field.
type EventIsResolutionFieldResolver interface {
	// IsResolution implements response to request for isResolution field.
	IsResolution(p graphql.ResolveParams) (bool, error)
}

// EventWasSilencedFieldResolver implement to resolve requests for the Event's wasSilenced field.
type EventWasSilencedFieldResolver interface {
	// WasSilenced implements response to request for wasSilenced field.
	WasSilenced(p graphql.ResolveParams) (bool, error)
}

// EventIsSilencedFieldResolver implement to resolve requests for the Event's isSilenced field.
type EventIsSilencedFieldResolver interface {
	// IsSilenced implements response to request for isSilenced field.
	IsSilenced(p graphql.ResolveParams) (bool, error)
}

// EventSilencesFieldResolver implement to resolve requests for the Event's silences field.
type EventSilencesFieldResolver interface {
	// Silences implements response to request for silences field.
	Silences(p graphql.ResolveParams) (interface{}, error)
}

// EventSilencedFieldResolver implement to resolve requests for the Event's silenced field.
type EventSilencedFieldResolver interface {
	// Silenced implements response to request for silenced field.
	Silenced(p graphql.ResolveParams) ([]string, error)
}

// EventToJSONFieldResolver implement to resolve requests for the Event's toJSON field.
type EventToJSONFieldResolver interface {
	// ToJSON implements response to request for toJSON field.
	ToJSON(p graphql.ResolveParams) (interface{}, error)
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
	EventMetadataFieldResolver
	EventTimestampFieldResolver
	EventEntityFieldResolver
	EventCheckFieldResolver
	EventHooksFieldResolver
	EventIsIncidentFieldResolver
	EventIsNewIncidentFieldResolver
	EventIsResolutionFieldResolver
	EventWasSilencedFieldResolver
	EventIsSilencedFieldResolver
	EventSilencesFieldResolver
	EventSilencedFieldResolver
	EventToJSONFieldResolver
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
func (_ EventAliases) ID(p graphql.ResolveParams) (string, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret, ok := val.(string)
	if err != nil {
		return ret, err
	}
	if !ok {
		return ret, errors.New("unable to coerce value for field 'id'")
	}
	return ret, err
}

// Namespace implements response to request for 'namespace' field.
func (_ EventAliases) Namespace(p graphql.ResolveParams) (string, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret, ok := val.(string)
	if err != nil {
		return ret, err
	}
	if !ok {
		return ret, errors.New("unable to coerce value for field 'namespace'")
	}
	return ret, err
}

// Metadata implements response to request for 'metadata' field.
func (_ EventAliases) Metadata(p graphql.ResolveParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// Timestamp implements response to request for 'timestamp' field.
func (_ EventAliases) Timestamp(p graphql.ResolveParams) (time.Time, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret, ok := val.(time.Time)
	if err != nil {
		return ret, err
	}
	if !ok {
		return ret, errors.New("unable to coerce value for field 'timestamp'")
	}
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
	ret, ok := val.(bool)
	if err != nil {
		return ret, err
	}
	if !ok {
		return ret, errors.New("unable to coerce value for field 'isIncident'")
	}
	return ret, err
}

// IsNewIncident implements response to request for 'isNewIncident' field.
func (_ EventAliases) IsNewIncident(p graphql.ResolveParams) (bool, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret, ok := val.(bool)
	if err != nil {
		return ret, err
	}
	if !ok {
		return ret, errors.New("unable to coerce value for field 'isNewIncident'")
	}
	return ret, err
}

// IsResolution implements response to request for 'isResolution' field.
func (_ EventAliases) IsResolution(p graphql.ResolveParams) (bool, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret, ok := val.(bool)
	if err != nil {
		return ret, err
	}
	if !ok {
		return ret, errors.New("unable to coerce value for field 'isResolution'")
	}
	return ret, err
}

// WasSilenced implements response to request for 'wasSilenced' field.
func (_ EventAliases) WasSilenced(p graphql.ResolveParams) (bool, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret, ok := val.(bool)
	if err != nil {
		return ret, err
	}
	if !ok {
		return ret, errors.New("unable to coerce value for field 'wasSilenced'")
	}
	return ret, err
}

// IsSilenced implements response to request for 'isSilenced' field.
func (_ EventAliases) IsSilenced(p graphql.ResolveParams) (bool, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret, ok := val.(bool)
	if err != nil {
		return ret, err
	}
	if !ok {
		return ret, errors.New("unable to coerce value for field 'isSilenced'")
	}
	return ret, err
}

// Silences implements response to request for 'silences' field.
func (_ EventAliases) Silences(p graphql.ResolveParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// Silenced implements response to request for 'silenced' field.
func (_ EventAliases) Silenced(p graphql.ResolveParams) ([]string, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret, ok := val.([]string)
	if err != nil {
		return ret, err
	}
	if !ok {
		return ret, errors.New("unable to coerce value for field 'silenced'")
	}
	return ret, err
}

// ToJSON implements response to request for 'toJSON' field.
func (_ EventAliases) ToJSON(p graphql.ResolveParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// EventType An Event is the encapsulating type sent across the Sensu websocket transport.
var EventType = graphql.NewType("Event", graphql.ObjectKind)

// RegisterEvent registers Event object type with given service.
func RegisterEvent(svc *graphql.Service, impl EventFieldResolvers) {
	svc.RegisterObject(_ObjectTypeEventDesc, impl)
}
func _ObjTypeEventIDHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EventIDFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.ID(frp)
	}
}

func _ObjTypeEventNamespaceHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EventNamespaceFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Namespace(frp)
	}
}

func _ObjTypeEventMetadataHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EventMetadataFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Metadata(frp)
	}
}

func _ObjTypeEventTimestampHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EventTimestampFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Timestamp(frp)
	}
}

func _ObjTypeEventEntityHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EventEntityFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Entity(frp)
	}
}

func _ObjTypeEventCheckHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EventCheckFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Check(frp)
	}
}

func _ObjTypeEventHooksHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EventHooksFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Hooks(frp)
	}
}

func _ObjTypeEventIsIncidentHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EventIsIncidentFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.IsIncident(frp)
	}
}

func _ObjTypeEventIsNewIncidentHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EventIsNewIncidentFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.IsNewIncident(frp)
	}
}

func _ObjTypeEventIsResolutionHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EventIsResolutionFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.IsResolution(frp)
	}
}

func _ObjTypeEventWasSilencedHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EventWasSilencedFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.WasSilenced(frp)
	}
}

func _ObjTypeEventIsSilencedHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EventIsSilencedFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.IsSilenced(frp)
	}
}

func _ObjTypeEventSilencesHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EventSilencesFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Silences(frp)
	}
}

func _ObjTypeEventSilencedHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EventSilencedFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Silenced(frp)
	}
}

func _ObjTypeEventToJSONHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EventToJSONFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.ToJSON(frp)
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
			"isNewIncident": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "isNewIncident returns true if the event is an incident but it's previous\niteration was OK.",
				Name:              "isNewIncident",
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
			"metadata": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "metadata contains name, namespace, labels and annotations of the record",
				Name:              "metadata",
				Type:              graphql1.NewNonNull(graphql.OutputType("ObjectMeta")),
			},
			"namespace": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "namespace in which this record resides",
				Name:              "namespace",
				Type:              graphql1.NewNonNull(graphql1.String),
			},
			"silenced": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "Silenced is a list of silenced entry ids (subscription and check name)",
				Name:              "silenced",
				Type:              graphql1.NewList(graphql1.String),
			},
			"silences": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "all current silences matching the check and entity's subscriptions.",
				Name:              "silences",
				Type:              graphql1.NewNonNull(graphql1.NewList(graphql1.NewNonNull(graphql.OutputType("Silenced")))),
			},
			"timestamp": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "Timestamp is the time in seconds since the Epoch.",
				Name:              "timestamp",
				Type:              graphql1.NewNonNull(graphql1.DateTime),
			},
			"toJSON": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "toJSON returns a REST API compatible representation of the resource. Handy for\nsharing snippets that can then be imported with `sensuctl create`.",
				Name:              "toJSON",
				Type:              graphql1.NewNonNull(graphql.OutputType("JSON")),
			},
			"wasSilenced": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "wasSilenced reflects whether the event was silenced when passing through the pipeline.",
				Name:              "wasSilenced",
				Type:              graphql1.NewNonNull(graphql1.Boolean),
			},
		},
		Interfaces: []*graphql1.Interface{
			graphql.Interface("Node"),
			graphql.Interface("Namespaced"),
			graphql.Interface("Silenceable"),
			graphql.Interface("Resource")},
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
		"check":         _ObjTypeEventCheckHandler,
		"entity":        _ObjTypeEventEntityHandler,
		"hooks":         _ObjTypeEventHooksHandler,
		"id":            _ObjTypeEventIDHandler,
		"isIncident":    _ObjTypeEventIsIncidentHandler,
		"isNewIncident": _ObjTypeEventIsNewIncidentHandler,
		"isResolution":  _ObjTypeEventIsResolutionHandler,
		"isSilenced":    _ObjTypeEventIsSilencedHandler,
		"metadata":      _ObjTypeEventMetadataHandler,
		"namespace":     _ObjTypeEventNamespaceHandler,
		"silenced":      _ObjTypeEventSilencedHandler,
		"silences":      _ObjTypeEventSilencesHandler,
		"timestamp":     _ObjTypeEventTimestampHandler,
		"toJSON":        _ObjTypeEventToJSONHandler,
		"wasSilenced":   _ObjTypeEventWasSilencedHandler,
	},
}

// EventConnectionNodesFieldResolver implement to resolve requests for the EventConnection's nodes field.
type EventConnectionNodesFieldResolver interface {
	// Nodes implements response to request for nodes field.
	Nodes(p graphql.ResolveParams) (interface{}, error)
}

// EventConnectionPageInfoFieldResolver implement to resolve requests for the EventConnection's pageInfo field.
type EventConnectionPageInfoFieldResolver interface {
	// PageInfo implements response to request for pageInfo field.
	PageInfo(p graphql.ResolveParams) (interface{}, error)
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
	EventConnectionNodesFieldResolver
	EventConnectionPageInfoFieldResolver
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

// Nodes implements response to request for 'nodes' field.
func (_ EventConnectionAliases) Nodes(p graphql.ResolveParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// PageInfo implements response to request for 'pageInfo' field.
func (_ EventConnectionAliases) PageInfo(p graphql.ResolveParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// EventConnectionType A connection to a sequence of records.
var EventConnectionType = graphql.NewType("EventConnection", graphql.ObjectKind)

// RegisterEventConnection registers EventConnection object type with given service.
func RegisterEventConnection(svc *graphql.Service, impl EventConnectionFieldResolvers) {
	svc.RegisterObject(_ObjectTypeEventConnectionDesc, impl)
}
func _ObjTypeEventConnectionNodesHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EventConnectionNodesFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Nodes(frp)
	}
}

func _ObjTypeEventConnectionPageInfoHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EventConnectionPageInfoFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.PageInfo(frp)
	}
}

func _ObjectTypeEventConnectionConfigFn() graphql1.ObjectConfig {
	return graphql1.ObjectConfig{
		Description: "A connection to a sequence of records.",
		Fields: graphql1.Fields{
			"nodes": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "self descriptive",
				Name:              "nodes",
				Type:              graphql1.NewNonNull(graphql1.NewList(graphql1.NewNonNull(graphql.OutputType("Event")))),
			},
			"pageInfo": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "self descriptive",
				Name:              "pageInfo",
				Type:              graphql1.NewNonNull(graphql.OutputType("OffsetPageInfo")),
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
		"nodes":    _ObjTypeEventConnectionNodesHandler,
		"pageInfo": _ObjTypeEventConnectionPageInfoHandler,
	},
}

// EventsListOrder Describes ways in which a list of events can be ordered.
type EventsListOrder string

// EventsListOrders holds enum values
var EventsListOrders = _EnumTypeEventsListOrderValues{
	LASTOK:   "LASTOK",
	NEWEST:   "NEWEST",
	OLDEST:   "OLDEST",
	SEVERITY: "SEVERITY",
}

// EventsListOrderType Describes ways in which a list of events can be ordered.
var EventsListOrderType = graphql.NewType("EventsListOrder", graphql.EnumKind)

// RegisterEventsListOrder registers EventsListOrder object type with given service.
func RegisterEventsListOrder(svc *graphql.Service) {
	svc.RegisterEnum(_EnumTypeEventsListOrderDesc)
}
func _EnumTypeEventsListOrderConfigFn() graphql1.EnumConfig {
	return graphql1.EnumConfig{
		Description: "Describes ways in which a list of events can be ordered.",
		Name:        "EventsListOrder",
		Values: graphql1.EnumValueConfigMap{
			"LASTOK": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "LASTOK",
			},
			"NEWEST": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "NEWEST",
			},
			"OLDEST": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "OLDEST",
			},
			"SEVERITY": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "SEVERITY",
			},
		},
	}
}

// describe EventsListOrder's configuration; kept private to avoid unintentional tampering of configuration at runtime.
var _EnumTypeEventsListOrderDesc = graphql.EnumDesc{Config: _EnumTypeEventsListOrderConfigFn}

type _EnumTypeEventsListOrderValues struct {
	// LASTOK - self descriptive
	LASTOK EventsListOrder
	// NEWEST - self descriptive
	NEWEST EventsListOrder
	// OLDEST - self descriptive
	OLDEST EventsListOrder
	// SEVERITY - self descriptive
	SEVERITY EventsListOrder
}
