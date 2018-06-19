// Code generated by scripts/gengraphql.go. DO NOT EDIT.

package schema

import (
	fmt "fmt"
	graphql1 "github.com/graphql-go/graphql"
	mapstructure "github.com/mitchellh/mapstructure"
	graphql "github.com/sensu/sensu-go/graphql"
)

// EnvironmentIDFieldResolver implement to resolve requests for the Environment's id field.
type EnvironmentIDFieldResolver interface {
	// ID implements response to request for id field.
	ID(p graphql.ResolveParams) (string, error)
}

// EnvironmentDescriptionFieldResolver implement to resolve requests for the Environment's description field.
type EnvironmentDescriptionFieldResolver interface {
	// Description implements response to request for description field.
	Description(p graphql.ResolveParams) (string, error)
}

// EnvironmentNameFieldResolver implement to resolve requests for the Environment's name field.
type EnvironmentNameFieldResolver interface {
	// Name implements response to request for name field.
	Name(p graphql.ResolveParams) (string, error)
}

// EnvironmentColourIDFieldResolver implement to resolve requests for the Environment's colourId field.
type EnvironmentColourIDFieldResolver interface {
	// ColourID implements response to request for colourId field.
	ColourID(p graphql.ResolveParams) (MutedColour, error)
}

// EnvironmentOrganizationFieldResolver implement to resolve requests for the Environment's organization field.
type EnvironmentOrganizationFieldResolver interface {
	// Organization implements response to request for organization field.
	Organization(p graphql.ResolveParams) (interface{}, error)
}

// EnvironmentChecksFieldResolverArgs contains arguments provided to checks when selected
type EnvironmentChecksFieldResolverArgs struct {
	Offset  int            // Offset - self descriptive
	Limit   int            // Limit adds optional limit to the number of entries returned.
	OrderBy CheckListOrder // OrderBy adds optional order to the records retrieved.
	Filter  string         // Filter reduces the set using the given Sensu Query Expression predicate.
}

// EnvironmentChecksFieldResolverParams contains contextual info to resolve checks field
type EnvironmentChecksFieldResolverParams struct {
	graphql.ResolveParams
	Args EnvironmentChecksFieldResolverArgs
}

// EnvironmentChecksFieldResolver implement to resolve requests for the Environment's checks field.
type EnvironmentChecksFieldResolver interface {
	// Checks implements response to request for checks field.
	Checks(p EnvironmentChecksFieldResolverParams) (interface{}, error)
}

// EnvironmentEntitiesFieldResolverArgs contains arguments provided to entities when selected
type EnvironmentEntitiesFieldResolverArgs struct {
	Offset  int             // Offset - self descriptive
	Limit   int             // Limit adds optional limit to the number of entries returned.
	OrderBy EntityListOrder // OrderBy adds optional order to the records retrieved.
	Filter  string          // Filter reduces the set using the given Sensu Query Expression predicate.
}

// EnvironmentEntitiesFieldResolverParams contains contextual info to resolve entities field
type EnvironmentEntitiesFieldResolverParams struct {
	graphql.ResolveParams
	Args EnvironmentEntitiesFieldResolverArgs
}

// EnvironmentEntitiesFieldResolver implement to resolve requests for the Environment's entities field.
type EnvironmentEntitiesFieldResolver interface {
	// Entities implements response to request for entities field.
	Entities(p EnvironmentEntitiesFieldResolverParams) (interface{}, error)
}

// EnvironmentEventsFieldResolverArgs contains arguments provided to events when selected
type EnvironmentEventsFieldResolverArgs struct {
	Offset  int             // Offset - self descriptive
	Limit   int             // Limit adds optional limit to the number of entries returned.
	OrderBy EventsListOrder // OrderBy adds optional order to the records retrieved.
	Filter  string          // Filter reduces the set using the given Sensu Query Expression predicate.
}

// EnvironmentEventsFieldResolverParams contains contextual info to resolve events field
type EnvironmentEventsFieldResolverParams struct {
	graphql.ResolveParams
	Args EnvironmentEventsFieldResolverArgs
}

// EnvironmentEventsFieldResolver implement to resolve requests for the Environment's events field.
type EnvironmentEventsFieldResolver interface {
	// Events implements response to request for events field.
	Events(p EnvironmentEventsFieldResolverParams) (interface{}, error)
}

// EnvironmentSilencesFieldResolverArgs contains arguments provided to silences when selected
type EnvironmentSilencesFieldResolverArgs struct {
	Offset int // Offset - self descriptive
	Limit  int // Limit adds optional limit to the number of entries returned.
}

// EnvironmentSilencesFieldResolverParams contains contextual info to resolve silences field
type EnvironmentSilencesFieldResolverParams struct {
	graphql.ResolveParams
	Args EnvironmentSilencesFieldResolverArgs
}

// EnvironmentSilencesFieldResolver implement to resolve requests for the Environment's silences field.
type EnvironmentSilencesFieldResolver interface {
	// Silences implements response to request for silences field.
	Silences(p EnvironmentSilencesFieldResolverParams) (interface{}, error)
}

// EnvironmentSubscriptionsFieldResolverArgs contains arguments provided to subscriptions when selected
type EnvironmentSubscriptionsFieldResolverArgs struct {
	OmitEntity bool                 // OmitEntity - Omit entity subscriptions from set.
	OrderBy    SubscriptionSetOrder // OrderBy adds optional order to the records retrieved.
}

// EnvironmentSubscriptionsFieldResolverParams contains contextual info to resolve subscriptions field
type EnvironmentSubscriptionsFieldResolverParams struct {
	graphql.ResolveParams
	Args EnvironmentSubscriptionsFieldResolverArgs
}

// EnvironmentSubscriptionsFieldResolver implement to resolve requests for the Environment's subscriptions field.
type EnvironmentSubscriptionsFieldResolver interface {
	// Subscriptions implements response to request for subscriptions field.
	Subscriptions(p EnvironmentSubscriptionsFieldResolverParams) (interface{}, error)
}

// EnvironmentCheckHistoryFieldResolverArgs contains arguments provided to checkHistory when selected
type EnvironmentCheckHistoryFieldResolverArgs struct {
	Filter string // Filter reduces the set using the given Sensu Query Expression predicate.
	Limit  int    // Limit adds optional limit to the number of entries returned.
}

// EnvironmentCheckHistoryFieldResolverParams contains contextual info to resolve checkHistory field
type EnvironmentCheckHistoryFieldResolverParams struct {
	graphql.ResolveParams
	Args EnvironmentCheckHistoryFieldResolverArgs
}

// EnvironmentCheckHistoryFieldResolver implement to resolve requests for the Environment's checkHistory field.
type EnvironmentCheckHistoryFieldResolver interface {
	// CheckHistory implements response to request for checkHistory field.
	CheckHistory(p EnvironmentCheckHistoryFieldResolverParams) (interface{}, error)
}

//
// EnvironmentFieldResolvers represents a collection of methods whose products represent the
// response values of the 'Environment' type.
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
type EnvironmentFieldResolvers interface {
	EnvironmentIDFieldResolver
	EnvironmentDescriptionFieldResolver
	EnvironmentNameFieldResolver
	EnvironmentColourIDFieldResolver
	EnvironmentOrganizationFieldResolver
	EnvironmentChecksFieldResolver
	EnvironmentEntitiesFieldResolver
	EnvironmentEventsFieldResolver
	EnvironmentSilencesFieldResolver
	EnvironmentSubscriptionsFieldResolver
	EnvironmentCheckHistoryFieldResolver
}

// EnvironmentAliases implements all methods on EnvironmentFieldResolvers interface by using reflection to
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
type EnvironmentAliases struct{}

// ID implements response to request for 'id' field.
func (_ EnvironmentAliases) ID(p graphql.ResolveParams) (string, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret := fmt.Sprint(val)
	return ret, err
}

// Description implements response to request for 'description' field.
func (_ EnvironmentAliases) Description(p graphql.ResolveParams) (string, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret := fmt.Sprint(val)
	return ret, err
}

// Name implements response to request for 'name' field.
func (_ EnvironmentAliases) Name(p graphql.ResolveParams) (string, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret := fmt.Sprint(val)
	return ret, err
}

// ColourID implements response to request for 'colourId' field.
func (_ EnvironmentAliases) ColourID(p graphql.ResolveParams) (MutedColour, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret := MutedColour(val.(string))
	return ret, err
}

// Organization implements response to request for 'organization' field.
func (_ EnvironmentAliases) Organization(p graphql.ResolveParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// Checks implements response to request for 'checks' field.
func (_ EnvironmentAliases) Checks(p EnvironmentChecksFieldResolverParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// Entities implements response to request for 'entities' field.
func (_ EnvironmentAliases) Entities(p EnvironmentEntitiesFieldResolverParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// Events implements response to request for 'events' field.
func (_ EnvironmentAliases) Events(p EnvironmentEventsFieldResolverParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// Silences implements response to request for 'silences' field.
func (_ EnvironmentAliases) Silences(p EnvironmentSilencesFieldResolverParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// Subscriptions implements response to request for 'subscriptions' field.
func (_ EnvironmentAliases) Subscriptions(p EnvironmentSubscriptionsFieldResolverParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// CheckHistory implements response to request for 'checkHistory' field.
func (_ EnvironmentAliases) CheckHistory(p EnvironmentCheckHistoryFieldResolverParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// EnvironmentType Environment represents a Sensu environment in RBAC
var EnvironmentType = graphql.NewType("Environment", graphql.ObjectKind)

// RegisterEnvironment registers Environment object type with given service.
func RegisterEnvironment(svc *graphql.Service, impl EnvironmentFieldResolvers) {
	svc.RegisterObject(_ObjectTypeEnvironmentDesc, impl)
}
func _ObjTypeEnvironmentIDHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EnvironmentIDFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.ID(frp)
	}
}

func _ObjTypeEnvironmentDescriptionHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EnvironmentDescriptionFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Description(frp)
	}
}

func _ObjTypeEnvironmentNameHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EnvironmentNameFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Name(frp)
	}
}

func _ObjTypeEnvironmentColourIDHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EnvironmentColourIDFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {

		val, err := resolver.ColourID(frp)
		return string(val), err
	}
}

func _ObjTypeEnvironmentOrganizationHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EnvironmentOrganizationFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Organization(frp)
	}
}

func _ObjTypeEnvironmentChecksHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EnvironmentChecksFieldResolver)
	return func(p graphql1.ResolveParams) (interface{}, error) {
		frp := EnvironmentChecksFieldResolverParams{ResolveParams: p}
		err := mapstructure.Decode(p.Args, &frp.Args)
		if err != nil {
			return nil, err
		}

		return resolver.Checks(frp)
	}
}

func _ObjTypeEnvironmentEntitiesHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EnvironmentEntitiesFieldResolver)
	return func(p graphql1.ResolveParams) (interface{}, error) {
		frp := EnvironmentEntitiesFieldResolverParams{ResolveParams: p}
		err := mapstructure.Decode(p.Args, &frp.Args)
		if err != nil {
			return nil, err
		}

		return resolver.Entities(frp)
	}
}

func _ObjTypeEnvironmentEventsHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EnvironmentEventsFieldResolver)
	return func(p graphql1.ResolveParams) (interface{}, error) {
		frp := EnvironmentEventsFieldResolverParams{ResolveParams: p}
		err := mapstructure.Decode(p.Args, &frp.Args)
		if err != nil {
			return nil, err
		}

		return resolver.Events(frp)
	}
}

func _ObjTypeEnvironmentSilencesHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EnvironmentSilencesFieldResolver)
	return func(p graphql1.ResolveParams) (interface{}, error) {
		frp := EnvironmentSilencesFieldResolverParams{ResolveParams: p}
		err := mapstructure.Decode(p.Args, &frp.Args)
		if err != nil {
			return nil, err
		}

		return resolver.Silences(frp)
	}
}

func _ObjTypeEnvironmentSubscriptionsHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EnvironmentSubscriptionsFieldResolver)
	return func(p graphql1.ResolveParams) (interface{}, error) {
		frp := EnvironmentSubscriptionsFieldResolverParams{ResolveParams: p}
		err := mapstructure.Decode(p.Args, &frp.Args)
		if err != nil {
			return nil, err
		}

		return resolver.Subscriptions(frp)
	}
}

func _ObjTypeEnvironmentCheckHistoryHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(EnvironmentCheckHistoryFieldResolver)
	return func(p graphql1.ResolveParams) (interface{}, error) {
		frp := EnvironmentCheckHistoryFieldResolverParams{ResolveParams: p}
		err := mapstructure.Decode(p.Args, &frp.Args)
		if err != nil {
			return nil, err
		}

		return resolver.CheckHistory(frp)
	}
}

func _ObjectTypeEnvironmentConfigFn() graphql1.ObjectConfig {
	return graphql1.ObjectConfig{
		Description: "Environment represents a Sensu environment in RBAC",
		Fields: graphql1.Fields{
			"checkHistory": &graphql1.Field{
				Args: graphql1.FieldConfigArgument{
					"filter": &graphql1.ArgumentConfig{
						DefaultValue: "",
						Description:  "Filter reduces the set using the given Sensu Query Expression predicate.",
						Type:         graphql1.String,
					},
					"limit": &graphql1.ArgumentConfig{
						DefaultValue: 10000,
						Description:  "Limit adds optional limit to the number of entries returned.",
						Type:         graphql1.Int,
					},
				},
				DeprecationReason: "",
				Description:       "checkHistory includes all persisted check execution results associated with\nthe environment. Unlike the Check type's history this field includes the most\nrecent result.",
				Name:              "checkHistory",
				Type:              graphql1.NewNonNull(graphql1.NewList(graphql.OutputType("CheckHistory"))),
			},
			"checks": &graphql1.Field{
				Args: graphql1.FieldConfigArgument{
					"filter": &graphql1.ArgumentConfig{
						DefaultValue: "",
						Description:  "Filter reduces the set using the given Sensu Query Expression predicate.",
						Type:         graphql1.String,
					},
					"limit": &graphql1.ArgumentConfig{
						DefaultValue: 10,
						Description:  "Limit adds optional limit to the number of entries returned.",
						Type:         graphql1.Int,
					},
					"offset": &graphql1.ArgumentConfig{
						DefaultValue: 0,
						Description:  "self descriptive",
						Type:         graphql1.Int,
					},
					"orderBy": &graphql1.ArgumentConfig{
						DefaultValue: "NAME_DESC",
						Description:  "OrderBy adds optional order to the records retrieved.",
						Type:         graphql.InputType("CheckListOrder"),
					},
				},
				DeprecationReason: "",
				Description:       "All check configurations associated with the environment.",
				Name:              "checks",
				Type:              graphql1.NewNonNull(graphql.OutputType("CheckConfigConnection")),
			},
			"colourId": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "ColourId. Experimental. Use graphical interfaces as symbolic reference to environment",
				Name:              "colourId",
				Type:              graphql1.NewNonNull(graphql.OutputType("MutedColour")),
			},
			"description": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "The description given to the environment.",
				Name:              "description",
				Type:              graphql1.String,
			},
			"entities": &graphql1.Field{
				Args: graphql1.FieldConfigArgument{
					"filter": &graphql1.ArgumentConfig{
						DefaultValue: "",
						Description:  "Filter reduces the set using the given Sensu Query Expression predicate.",
						Type:         graphql1.String,
					},
					"limit": &graphql1.ArgumentConfig{
						DefaultValue: 10,
						Description:  "Limit adds optional limit to the number of entries returned.",
						Type:         graphql1.Int,
					},
					"offset": &graphql1.ArgumentConfig{
						DefaultValue: 0,
						Description:  "self descriptive",
						Type:         graphql1.Int,
					},
					"orderBy": &graphql1.ArgumentConfig{
						DefaultValue: "ID_DESC",
						Description:  "OrderBy adds optional order to the records retrieved.",
						Type:         graphql.InputType("EntityListOrder"),
					},
				},
				DeprecationReason: "",
				Description:       "All entities associated with the environment.",
				Name:              "entities",
				Type:              graphql1.NewNonNull(graphql.OutputType("EntityConnection")),
			},
			"events": &graphql1.Field{
				Args: graphql1.FieldConfigArgument{
					"filter": &graphql1.ArgumentConfig{
						DefaultValue: "",
						Description:  "Filter reduces the set using the given Sensu Query Expression predicate.",
						Type:         graphql1.String,
					},
					"limit": &graphql1.ArgumentConfig{
						DefaultValue: 10,
						Description:  "Limit adds optional limit to the number of entries returned.",
						Type:         graphql1.Int,
					},
					"offset": &graphql1.ArgumentConfig{
						DefaultValue: 0,
						Description:  "self descriptive",
						Type:         graphql1.Int,
					},
					"orderBy": &graphql1.ArgumentConfig{
						DefaultValue: "SEVERITY",
						Description:  "OrderBy adds optional order to the records retrieved.",
						Type:         graphql.InputType("EventsListOrder"),
					},
				},
				DeprecationReason: "",
				Description:       "All events associated with the environment.",
				Name:              "events",
				Type:              graphql1.NewNonNull(graphql.OutputType("EventConnection")),
			},
			"id": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "The globally unique identifier of the record.",
				Name:              "id",
				Type:              graphql1.NewNonNull(graphql1.ID),
			},
			"name": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "name is the unique identifier for a organization.",
				Name:              "name",
				Type:              graphql1.NewNonNull(graphql1.String),
			},
			"organization": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "The organization the environment belongs to.",
				Name:              "organization",
				Type:              graphql1.NewNonNull(graphql.OutputType("Organization")),
			},
			"silences": &graphql1.Field{
				Args: graphql1.FieldConfigArgument{
					"limit": &graphql1.ArgumentConfig{
						DefaultValue: 10,
						Description:  "Limit adds optional limit to the number of entries returned.",
						Type:         graphql1.Int,
					},
					"offset": &graphql1.ArgumentConfig{
						DefaultValue: 0,
						Description:  "self descriptive",
						Type:         graphql1.Int,
					},
				},
				DeprecationReason: "",
				Description:       "All silences associated with the environment.",
				Name:              "silences",
				Type:              graphql1.NewNonNull(graphql.OutputType("SilencedConnection")),
			},
			"subscriptions": &graphql1.Field{
				Args: graphql1.FieldConfigArgument{
					"omitEntity": &graphql1.ArgumentConfig{
						DefaultValue: false,
						Description:  "Omit entity subscriptions from set.",
						Type:         graphql1.Boolean,
					},
					"orderBy": &graphql1.ArgumentConfig{
						DefaultValue: "OCCURRENCES",
						Description:  "OrderBy adds optional order to the records retrieved.",
						Type:         graphql.InputType("SubscriptionSetOrder"),
					},
				},
				DeprecationReason: "",
				Description:       "All subscriptions in use in the environment.",
				Name:              "subscriptions",
				Type:              graphql1.NewNonNull(graphql.OutputType("SubscriptionSet")),
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
			panic("Unimplemented; see EnvironmentFieldResolvers.")
		},
		Name: "Environment",
	}
}

// describe Environment's configuration; kept private to avoid unintentional tampering of configuration at runtime.
var _ObjectTypeEnvironmentDesc = graphql.ObjectDesc{
	Config: _ObjectTypeEnvironmentConfigFn,
	FieldHandlers: map[string]graphql.FieldHandler{
		"checkHistory":  _ObjTypeEnvironmentCheckHistoryHandler,
		"checks":        _ObjTypeEnvironmentChecksHandler,
		"colourId":      _ObjTypeEnvironmentColourIDHandler,
		"description":   _ObjTypeEnvironmentDescriptionHandler,
		"entities":      _ObjTypeEnvironmentEntitiesHandler,
		"events":        _ObjTypeEnvironmentEventsHandler,
		"id":            _ObjTypeEnvironmentIDHandler,
		"name":          _ObjTypeEnvironmentNameHandler,
		"organization":  _ObjTypeEnvironmentOrganizationHandler,
		"silences":      _ObjTypeEnvironmentSilencesHandler,
		"subscriptions": _ObjTypeEnvironmentSubscriptionsHandler,
	},
}

// SubscriptionSetOrder Describes ways in which a set of subscriptions can be ordered.
type SubscriptionSetOrder string

// SubscriptionSetOrders holds enum values
var SubscriptionSetOrders = _EnumTypeSubscriptionSetOrderValues{
	ALPHA_ASC:   "ALPHA_ASC",
	ALPHA_DESC:  "ALPHA_DESC",
	OCCURRENCES: "OCCURRENCES",
}

// SubscriptionSetOrderType Describes ways in which a set of subscriptions can be ordered.
var SubscriptionSetOrderType = graphql.NewType("SubscriptionSetOrder", graphql.EnumKind)

// RegisterSubscriptionSetOrder registers SubscriptionSetOrder object type with given service.
func RegisterSubscriptionSetOrder(svc *graphql.Service) {
	svc.RegisterEnum(_EnumTypeSubscriptionSetOrderDesc)
}
func _EnumTypeSubscriptionSetOrderConfigFn() graphql1.EnumConfig {
	return graphql1.EnumConfig{
		Description: "Describes ways in which a set of subscriptions can be ordered.",
		Name:        "SubscriptionSetOrder",
		Values: graphql1.EnumValueConfigMap{
			"ALPHA_ASC": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "ALPHA_ASC",
			},
			"ALPHA_DESC": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "ALPHA_DESC",
			},
			"OCCURRENCES": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "OCCURRENCES",
			},
		},
	}
}

// describe SubscriptionSetOrder's configuration; kept private to avoid unintentional tampering of configuration at runtime.
var _EnumTypeSubscriptionSetOrderDesc = graphql.EnumDesc{Config: _EnumTypeSubscriptionSetOrderConfigFn}

type _EnumTypeSubscriptionSetOrderValues struct {
	// ALPHA_ASC - self descriptive
	ALPHA_ASC SubscriptionSetOrder
	// ALPHA_DESC - self descriptive
	ALPHA_DESC SubscriptionSetOrder
	// OCCURRENCES - self descriptive
	OCCURRENCES SubscriptionSetOrder
}

// CheckListOrder self descriptive
type CheckListOrder string

// CheckListOrders holds enum values
var CheckListOrders = _EnumTypeCheckListOrderValues{
	NAME:      "NAME",
	NAME_DESC: "NAME_DESC",
}

// CheckListOrderType self descriptive
var CheckListOrderType = graphql.NewType("CheckListOrder", graphql.EnumKind)

// RegisterCheckListOrder registers CheckListOrder object type with given service.
func RegisterCheckListOrder(svc *graphql.Service) {
	svc.RegisterEnum(_EnumTypeCheckListOrderDesc)
}
func _EnumTypeCheckListOrderConfigFn() graphql1.EnumConfig {
	return graphql1.EnumConfig{
		Description: "self descriptive",
		Name:        "CheckListOrder",
		Values: graphql1.EnumValueConfigMap{
			"NAME": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "NAME",
			},
			"NAME_DESC": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "NAME_DESC",
			},
		},
	}
}

// describe CheckListOrder's configuration; kept private to avoid unintentional tampering of configuration at runtime.
var _EnumTypeCheckListOrderDesc = graphql.EnumDesc{Config: _EnumTypeCheckListOrderConfigFn}

type _EnumTypeCheckListOrderValues struct {
	// NAME - self descriptive
	NAME CheckListOrder
	// NAME_DESC - self descriptive
	NAME_DESC CheckListOrder
}

// EntityListOrder self descriptive
type EntityListOrder string

// EntityListOrders holds enum values
var EntityListOrders = _EnumTypeEntityListOrderValues{
	ID:       "ID",
	ID_DESC:  "ID_DESC",
	LASTSEEN: "LASTSEEN",
}

// EntityListOrderType self descriptive
var EntityListOrderType = graphql.NewType("EntityListOrder", graphql.EnumKind)

// RegisterEntityListOrder registers EntityListOrder object type with given service.
func RegisterEntityListOrder(svc *graphql.Service) {
	svc.RegisterEnum(_EnumTypeEntityListOrderDesc)
}
func _EnumTypeEntityListOrderConfigFn() graphql1.EnumConfig {
	return graphql1.EnumConfig{
		Description: "self descriptive",
		Name:        "EntityListOrder",
		Values: graphql1.EnumValueConfigMap{
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
			"LASTSEEN": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "LASTSEEN",
			},
		},
	}
}

// describe EntityListOrder's configuration; kept private to avoid unintentional tampering of configuration at runtime.
var _EnumTypeEntityListOrderDesc = graphql.EnumDesc{Config: _EnumTypeEntityListOrderConfigFn}

type _EnumTypeEntityListOrderValues struct {
	// ID - self descriptive
	ID EntityListOrder
	// ID_DESC - self descriptive
	ID_DESC EntityListOrder
	// LASTSEEN - self descriptive
	LASTSEEN EntityListOrder
}

// EventsListOrder self descriptive
type EventsListOrder string

// EventsListOrders holds enum values
var EventsListOrders = _EnumTypeEventsListOrderValues{
	NEWEST:   "NEWEST",
	OLDEST:   "OLDEST",
	SEVERITY: "SEVERITY",
}

// EventsListOrderType self descriptive
var EventsListOrderType = graphql.NewType("EventsListOrder", graphql.EnumKind)

// RegisterEventsListOrder registers EventsListOrder object type with given service.
func RegisterEventsListOrder(svc *graphql.Service) {
	svc.RegisterEnum(_EnumTypeEventsListOrderDesc)
}
func _EnumTypeEventsListOrderConfigFn() graphql1.EnumConfig {
	return graphql1.EnumConfig{
		Description: "self descriptive",
		Name:        "EventsListOrder",
		Values: graphql1.EnumValueConfigMap{
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
	// OLDEST - self descriptive
	OLDEST EventsListOrder
	// NEWEST - self descriptive
	NEWEST EventsListOrder
	// SEVERITY - self descriptive
	SEVERITY EventsListOrder
}

// MutedColour self descriptive
type MutedColour string

// MutedColours holds enum values
var MutedColours = _EnumTypeMutedColourValues{
	BLUE:   "BLUE",
	GRAY:   "GRAY",
	GREEN:  "GREEN",
	ORANGE: "ORANGE",
	PINK:   "PINK",
	PURPLE: "PURPLE",
	YELLOW: "YELLOW",
}

// MutedColourType self descriptive
var MutedColourType = graphql.NewType("MutedColour", graphql.EnumKind)

// RegisterMutedColour registers MutedColour object type with given service.
func RegisterMutedColour(svc *graphql.Service) {
	svc.RegisterEnum(_EnumTypeMutedColourDesc)
}
func _EnumTypeMutedColourConfigFn() graphql1.EnumConfig {
	return graphql1.EnumConfig{
		Description: "self descriptive",
		Name:        "MutedColour",
		Values: graphql1.EnumValueConfigMap{
			"BLUE": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "BLUE",
			},
			"GRAY": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "GRAY",
			},
			"GREEN": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "GREEN",
			},
			"ORANGE": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "ORANGE",
			},
			"PINK": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "PINK",
			},
			"PURPLE": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "PURPLE",
			},
			"YELLOW": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "YELLOW",
			},
		},
	}
}

// describe MutedColour's configuration; kept private to avoid unintentional tampering of configuration at runtime.
var _EnumTypeMutedColourDesc = graphql.EnumDesc{Config: _EnumTypeMutedColourConfigFn}

type _EnumTypeMutedColourValues struct {
	// BLUE - self descriptive
	BLUE MutedColour
	// GRAY - self descriptive
	GRAY MutedColour
	// GREEN - self descriptive
	GREEN MutedColour
	// ORANGE - self descriptive
	ORANGE MutedColour
	// PINK - self descriptive
	PINK MutedColour
	// PURPLE - self descriptive
	PURPLE MutedColour
	// YELLOW - self descriptive
	YELLOW MutedColour
}
