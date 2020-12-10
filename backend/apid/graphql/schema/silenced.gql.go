// Code generated by scripts/gengraphql.go. DO NOT EDIT.

package schema

import (
	errors "errors"
	graphql1 "github.com/graphql-go/graphql"
	graphql "github.com/sensu/sensu-go/graphql"
	time "time"
)

// SilencedIDFieldResolver implement to resolve requests for the Silenced's id field.
type SilencedIDFieldResolver interface {
	// ID implements response to request for id field.
	ID(p graphql.ResolveParams) (string, error)
}

// SilencedNamespaceFieldResolver implement to resolve requests for the Silenced's namespace field.
type SilencedNamespaceFieldResolver interface {
	// Namespace implements response to request for namespace field.
	Namespace(p graphql.ResolveParams) (string, error)
}

// SilencedNameFieldResolver implement to resolve requests for the Silenced's name field.
type SilencedNameFieldResolver interface {
	// Name implements response to request for name field.
	Name(p graphql.ResolveParams) (string, error)
}

// SilencedMetadataFieldResolver implement to resolve requests for the Silenced's metadata field.
type SilencedMetadataFieldResolver interface {
	// Metadata implements response to request for metadata field.
	Metadata(p graphql.ResolveParams) (interface{}, error)
}

// SilencedExpireFieldResolver implement to resolve requests for the Silenced's expire field.
type SilencedExpireFieldResolver interface {
	// Expire implements response to request for expire field.
	Expire(p graphql.ResolveParams) (int, error)
}

// SilencedExpiresFieldResolver implement to resolve requests for the Silenced's expires field.
type SilencedExpiresFieldResolver interface {
	// Expires implements response to request for expires field.
	Expires(p graphql.ResolveParams) (*time.Time, error)
}

// SilencedExpireOnResolveFieldResolver implement to resolve requests for the Silenced's expireOnResolve field.
type SilencedExpireOnResolveFieldResolver interface {
	// ExpireOnResolve implements response to request for expireOnResolve field.
	ExpireOnResolve(p graphql.ResolveParams) (bool, error)
}

// SilencedCreatorFieldResolver implement to resolve requests for the Silenced's creator field.
type SilencedCreatorFieldResolver interface {
	// Creator implements response to request for creator field.
	Creator(p graphql.ResolveParams) (string, error)
}

// SilencedCheckFieldResolver implement to resolve requests for the Silenced's check field.
type SilencedCheckFieldResolver interface {
	// Check implements response to request for check field.
	Check(p graphql.ResolveParams) (interface{}, error)
}

// SilencedReasonFieldResolver implement to resolve requests for the Silenced's reason field.
type SilencedReasonFieldResolver interface {
	// Reason implements response to request for reason field.
	Reason(p graphql.ResolveParams) (string, error)
}

// SilencedSubscriptionFieldResolver implement to resolve requests for the Silenced's subscription field.
type SilencedSubscriptionFieldResolver interface {
	// Subscription implements response to request for subscription field.
	Subscription(p graphql.ResolveParams) (string, error)
}

// SilencedBeginFieldResolver implement to resolve requests for the Silenced's begin field.
type SilencedBeginFieldResolver interface {
	// Begin implements response to request for begin field.
	Begin(p graphql.ResolveParams) (*time.Time, error)
}

// SilencedToJSONFieldResolver implement to resolve requests for the Silenced's toJSON field.
type SilencedToJSONFieldResolver interface {
	// ToJSON implements response to request for toJSON field.
	ToJSON(p graphql.ResolveParams) (interface{}, error)
}

//
// SilencedFieldResolvers represents a collection of methods whose products represent the
// response values of the 'Silenced' type.
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
type SilencedFieldResolvers interface {
	SilencedIDFieldResolver
	SilencedNamespaceFieldResolver
	SilencedNameFieldResolver
	SilencedMetadataFieldResolver
	SilencedExpireFieldResolver
	SilencedExpiresFieldResolver
	SilencedExpireOnResolveFieldResolver
	SilencedCreatorFieldResolver
	SilencedCheckFieldResolver
	SilencedReasonFieldResolver
	SilencedSubscriptionFieldResolver
	SilencedBeginFieldResolver
	SilencedToJSONFieldResolver
}

// SilencedAliases implements all methods on SilencedFieldResolvers interface by using reflection to
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
type SilencedAliases struct{}

// ID implements response to request for 'id' field.
func (_ SilencedAliases) ID(p graphql.ResolveParams) (string, error) {
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
func (_ SilencedAliases) Namespace(p graphql.ResolveParams) (string, error) {
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

// Name implements response to request for 'name' field.
func (_ SilencedAliases) Name(p graphql.ResolveParams) (string, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret, ok := val.(string)
	if err != nil {
		return ret, err
	}
	if !ok {
		return ret, errors.New("unable to coerce value for field 'name'")
	}
	return ret, err
}

// Metadata implements response to request for 'metadata' field.
func (_ SilencedAliases) Metadata(p graphql.ResolveParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// Expire implements response to request for 'expire' field.
func (_ SilencedAliases) Expire(p graphql.ResolveParams) (int, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret, ok := graphql1.Int.ParseValue(val).(int)
	if err != nil {
		return ret, err
	}
	if !ok {
		return ret, errors.New("unable to coerce value for field 'expire'")
	}
	return ret, err
}

// Expires implements response to request for 'expires' field.
func (_ SilencedAliases) Expires(p graphql.ResolveParams) (*time.Time, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret, ok := val.(*time.Time)
	if err != nil {
		return ret, err
	}
	if !ok {
		return ret, errors.New("unable to coerce value for field 'expires'")
	}
	return ret, err
}

// ExpireOnResolve implements response to request for 'expireOnResolve' field.
func (_ SilencedAliases) ExpireOnResolve(p graphql.ResolveParams) (bool, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret, ok := val.(bool)
	if err != nil {
		return ret, err
	}
	if !ok {
		return ret, errors.New("unable to coerce value for field 'expireOnResolve'")
	}
	return ret, err
}

// Creator implements response to request for 'creator' field.
func (_ SilencedAliases) Creator(p graphql.ResolveParams) (string, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret, ok := val.(string)
	if err != nil {
		return ret, err
	}
	if !ok {
		return ret, errors.New("unable to coerce value for field 'creator'")
	}
	return ret, err
}

// Check implements response to request for 'check' field.
func (_ SilencedAliases) Check(p graphql.ResolveParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// Reason implements response to request for 'reason' field.
func (_ SilencedAliases) Reason(p graphql.ResolveParams) (string, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret, ok := val.(string)
	if err != nil {
		return ret, err
	}
	if !ok {
		return ret, errors.New("unable to coerce value for field 'reason'")
	}
	return ret, err
}

// Subscription implements response to request for 'subscription' field.
func (_ SilencedAliases) Subscription(p graphql.ResolveParams) (string, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret, ok := val.(string)
	if err != nil {
		return ret, err
	}
	if !ok {
		return ret, errors.New("unable to coerce value for field 'subscription'")
	}
	return ret, err
}

// Begin implements response to request for 'begin' field.
func (_ SilencedAliases) Begin(p graphql.ResolveParams) (*time.Time, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret, ok := val.(*time.Time)
	if err != nil {
		return ret, err
	}
	if !ok {
		return ret, errors.New("unable to coerce value for field 'begin'")
	}
	return ret, err
}

// ToJSON implements response to request for 'toJSON' field.
func (_ SilencedAliases) ToJSON(p graphql.ResolveParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// SilencedType Silenced is the representation of a silence entry.
var SilencedType = graphql.NewType("Silenced", graphql.ObjectKind)

// RegisterSilenced registers Silenced object type with given service.
func RegisterSilenced(svc *graphql.Service, impl SilencedFieldResolvers) {
	svc.RegisterObject(_ObjectTypeSilencedDesc, impl)
}
func _ObjTypeSilencedIDHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(SilencedIDFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.ID(frp)
	}
}

func _ObjTypeSilencedNamespaceHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(SilencedNamespaceFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Namespace(frp)
	}
}

func _ObjTypeSilencedNameHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(SilencedNameFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Name(frp)
	}
}

func _ObjTypeSilencedMetadataHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(SilencedMetadataFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Metadata(frp)
	}
}

func _ObjTypeSilencedExpireHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(SilencedExpireFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Expire(frp)
	}
}

func _ObjTypeSilencedExpiresHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(SilencedExpiresFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Expires(frp)
	}
}

func _ObjTypeSilencedExpireOnResolveHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(SilencedExpireOnResolveFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.ExpireOnResolve(frp)
	}
}

func _ObjTypeSilencedCreatorHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(SilencedCreatorFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Creator(frp)
	}
}

func _ObjTypeSilencedCheckHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(SilencedCheckFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Check(frp)
	}
}

func _ObjTypeSilencedReasonHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(SilencedReasonFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Reason(frp)
	}
}

func _ObjTypeSilencedSubscriptionHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(SilencedSubscriptionFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Subscription(frp)
	}
}

func _ObjTypeSilencedBeginHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(SilencedBeginFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Begin(frp)
	}
}

func _ObjTypeSilencedToJSONHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(SilencedToJSONFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.ToJSON(frp)
	}
}

func _ObjectTypeSilencedConfigFn() graphql1.ObjectConfig {
	return graphql1.ObjectConfig{
		Description: "Silenced is the representation of a silence entry.",
		Fields: graphql1.Fields{
			"begin": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "Begin is a timestamp at which the silenced entry takes effect.",
				Name:              "begin",
				Type:              graphql1.DateTime,
			},
			"check": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "Check is the name of the check event to be silenced.",
				Name:              "check",
				Type:              graphql.OutputType("CheckConfig"),
			},
			"creator": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "Creator is the author of the silenced entry",
				Name:              "creator",
				Type:              graphql1.NewNonNull(graphql1.String),
			},
			"expire": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "Expire is the number of seconds the entry will live",
				Name:              "expire",
				Type:              graphql1.NewNonNull(graphql1.Int),
			},
			"expireOnResolve": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "ExpireOnResolve defaults to false, clears the entry on resolution when set\nto true",
				Name:              "expireOnResolve",
				Type:              graphql1.NewNonNull(graphql1.Boolean),
			},
			"expires": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "Exact time at which the silenced entry will expire",
				Name:              "expires",
				Type:              graphql1.DateTime,
			},
			"id": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "The globally unique identifier for the record.",
				Name:              "id",
				Type:              graphql1.NewNonNull(graphql1.ID),
			},
			"metadata": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "metadata contains name, namespace, labels and annotations of the record",
				Name:              "metadata",
				Type:              graphql.OutputType("ObjectMeta"),
			},
			"name": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "use metadata",
				Description:       "Name is the combination of subscription and check name (subscription:checkname)",
				Name:              "name",
				Type:              graphql1.NewNonNull(graphql1.String),
			},
			"namespace": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "use metadata",
				Description:       "The namespace the object belongs to.",
				Name:              "namespace",
				Type:              graphql1.NewNonNull(graphql1.String),
			},
			"reason": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "Reason is used to provide context to the entry",
				Name:              "reason",
				Type:              graphql1.String,
			},
			"subscription": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "Subscription is the name of the subscription to which the entry applies.",
				Name:              "subscription",
				Type:              graphql1.String,
			},
			"toJSON": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "toJSON returns a REST API compatible representation of the resource. Handy for\nsharing snippets that can then be imported with `sensuctl create`.",
				Name:              "toJSON",
				Type:              graphql1.NewNonNull(graphql.OutputType("JSON")),
			},
		},
		Interfaces: []*graphql1.Interface{
			graphql.Interface("Node"),
			graphql.Interface("Namespaced"),
			graphql.Interface("Resource")},
		IsTypeOf: func(_ graphql1.IsTypeOfParams) bool {
			// NOTE:
			// Panic by default. Intent is that when Service is invoked, values of
			// these fields are updated with instantiated resolvers. If these
			// defaults are called it is most certainly programmer err.
			// If you're see this comment then: 'Whoops! Sorry, my bad.'
			panic("Unimplemented; see SilencedFieldResolvers.")
		},
		Name: "Silenced",
	}
}

// describe Silenced's configuration; kept private to avoid unintentional tampering of configuration at runtime.
var _ObjectTypeSilencedDesc = graphql.ObjectDesc{
	Config: _ObjectTypeSilencedConfigFn,
	FieldHandlers: map[string]graphql.FieldHandler{
		"begin":           _ObjTypeSilencedBeginHandler,
		"check":           _ObjTypeSilencedCheckHandler,
		"creator":         _ObjTypeSilencedCreatorHandler,
		"expire":          _ObjTypeSilencedExpireHandler,
		"expireOnResolve": _ObjTypeSilencedExpireOnResolveHandler,
		"expires":         _ObjTypeSilencedExpiresHandler,
		"id":              _ObjTypeSilencedIDHandler,
		"metadata":        _ObjTypeSilencedMetadataHandler,
		"name":            _ObjTypeSilencedNameHandler,
		"namespace":       _ObjTypeSilencedNamespaceHandler,
		"reason":          _ObjTypeSilencedReasonHandler,
		"subscription":    _ObjTypeSilencedSubscriptionHandler,
		"toJSON":          _ObjTypeSilencedToJSONHandler,
	},
}

// SilenceableType Silenceable describes resources that can be silenced
var SilenceableType = graphql.NewType("Silenceable", graphql.InterfaceKind)

// RegisterSilenceable registers Silenceable object type with given service.
func RegisterSilenceable(svc *graphql.Service, impl graphql.InterfaceTypeResolver) {
	svc.RegisterInterface(_InterfaceTypeSilenceableDesc, impl)
}
func _InterfaceTypeSilenceableConfigFn() graphql1.InterfaceConfig {
	return graphql1.InterfaceConfig{
		Description: "Silenceable describes resources that can be silenced",
		Fields: graphql1.Fields{
			"isSilenced": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "self descriptive",
				Name:              "isSilenced",
				Type:              graphql1.NewNonNull(graphql1.Boolean),
			},
			"silences": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "self descriptive",
				Name:              "silences",
				Type:              graphql1.NewNonNull(graphql1.NewList(graphql1.NewNonNull(graphql.OutputType("Silenced")))),
			},
		},
		Name: "Silenceable",
		ResolveType: func(_ graphql1.ResolveTypeParams) *graphql1.Object {
			// NOTE:
			// Panic by default. Intent is that when Service is invoked, values of
			// these fields are updated with instantiated resolvers. If these
			// defaults are called it is most certainly programmer err.
			// If you're see this comment then: 'Whoops! Sorry, my bad.'
			panic("Unimplemented; see InterfaceTypeResolver.")
		},
	}
}

// describe Silenceable's configuration; kept private to avoid unintentional tampering of configuration at runtime.
var _InterfaceTypeSilenceableDesc = graphql.InterfaceDesc{Config: _InterfaceTypeSilenceableConfigFn}

// SilencedConnectionNodesFieldResolver implement to resolve requests for the SilencedConnection's nodes field.
type SilencedConnectionNodesFieldResolver interface {
	// Nodes implements response to request for nodes field.
	Nodes(p graphql.ResolveParams) (interface{}, error)
}

// SilencedConnectionPageInfoFieldResolver implement to resolve requests for the SilencedConnection's pageInfo field.
type SilencedConnectionPageInfoFieldResolver interface {
	// PageInfo implements response to request for pageInfo field.
	PageInfo(p graphql.ResolveParams) (interface{}, error)
}

//
// SilencedConnectionFieldResolvers represents a collection of methods whose products represent the
// response values of the 'SilencedConnection' type.
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
type SilencedConnectionFieldResolvers interface {
	SilencedConnectionNodesFieldResolver
	SilencedConnectionPageInfoFieldResolver
}

// SilencedConnectionAliases implements all methods on SilencedConnectionFieldResolvers interface by using reflection to
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
type SilencedConnectionAliases struct{}

// Nodes implements response to request for 'nodes' field.
func (_ SilencedConnectionAliases) Nodes(p graphql.ResolveParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// PageInfo implements response to request for 'pageInfo' field.
func (_ SilencedConnectionAliases) PageInfo(p graphql.ResolveParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// SilencedConnectionType A connection to a sequence of records.
var SilencedConnectionType = graphql.NewType("SilencedConnection", graphql.ObjectKind)

// RegisterSilencedConnection registers SilencedConnection object type with given service.
func RegisterSilencedConnection(svc *graphql.Service, impl SilencedConnectionFieldResolvers) {
	svc.RegisterObject(_ObjectTypeSilencedConnectionDesc, impl)
}
func _ObjTypeSilencedConnectionNodesHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(SilencedConnectionNodesFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Nodes(frp)
	}
}

func _ObjTypeSilencedConnectionPageInfoHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(SilencedConnectionPageInfoFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.PageInfo(frp)
	}
}

func _ObjectTypeSilencedConnectionConfigFn() graphql1.ObjectConfig {
	return graphql1.ObjectConfig{
		Description: "A connection to a sequence of records.",
		Fields: graphql1.Fields{
			"nodes": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "self descriptive",
				Name:              "nodes",
				Type:              graphql1.NewNonNull(graphql1.NewList(graphql1.NewNonNull(graphql.OutputType("Silenced")))),
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
			panic("Unimplemented; see SilencedConnectionFieldResolvers.")
		},
		Name: "SilencedConnection",
	}
}

// describe SilencedConnection's configuration; kept private to avoid unintentional tampering of configuration at runtime.
var _ObjectTypeSilencedConnectionDesc = graphql.ObjectDesc{
	Config: _ObjectTypeSilencedConnectionConfigFn,
	FieldHandlers: map[string]graphql.FieldHandler{
		"nodes":    _ObjTypeSilencedConnectionNodesHandler,
		"pageInfo": _ObjTypeSilencedConnectionPageInfoHandler,
	},
}

// SilencesListOrder Describes ways in which a list of silences can be ordered.
type SilencesListOrder string

// SilencesListOrders holds enum values
var SilencesListOrders = _EnumTypeSilencesListOrderValues{
	BEGIN:      "BEGIN",
	BEGIN_DESC: "BEGIN_DESC",
	ID:         "ID",
	ID_DESC:    "ID_DESC",
}

// SilencesListOrderType Describes ways in which a list of silences can be ordered.
var SilencesListOrderType = graphql.NewType("SilencesListOrder", graphql.EnumKind)

// RegisterSilencesListOrder registers SilencesListOrder object type with given service.
func RegisterSilencesListOrder(svc *graphql.Service) {
	svc.RegisterEnum(_EnumTypeSilencesListOrderDesc)
}
func _EnumTypeSilencesListOrderConfigFn() graphql1.EnumConfig {
	return graphql1.EnumConfig{
		Description: "Describes ways in which a list of silences can be ordered.",
		Name:        "SilencesListOrder",
		Values: graphql1.EnumValueConfigMap{
			"BEGIN": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "BEGIN",
			},
			"BEGIN_DESC": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "BEGIN_DESC",
			},
			"ID": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "ID",
			},
			"ID_DESC": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "ID_DESC",
			},
		},
	}
}

// describe SilencesListOrder's configuration; kept private to avoid unintentional tampering of configuration at runtime.
var _EnumTypeSilencesListOrderDesc = graphql.EnumDesc{Config: _EnumTypeSilencesListOrderConfigFn}

type _EnumTypeSilencesListOrderValues struct {
	// ID - self descriptive
	ID SilencesListOrder
	// ID_DESC - self descriptive
	ID_DESC SilencesListOrder
	// BEGIN - self descriptive
	BEGIN SilencesListOrder
	// BEGIN_DESC - self descriptive
	BEGIN_DESC SilencesListOrder
}
