// Code generated by scripts/gengraphql.go. DO NOT EDIT.

package schema

import (
	fmt "fmt"
	graphql1 "github.com/graphql-go/graphql"
	graphql "github.com/sensu/sensu-go/graphql"
)

// RuleNamespaceFieldResolver implement to resolve requests for the Rule's namespace field.
type RuleNamespaceFieldResolver interface {
	// Namespace implements response to request for namespace field.
	Namespace(p graphql.ResolveParams) (interface{}, error)
}

// RuleTypeFieldResolver implement to resolve requests for the Rule's type field.
type RuleTypeFieldResolver interface {
	// Type implements response to request for type field.
	Type(p graphql.ResolveParams) (RuleResource, error)
}

// RulePermissionsFieldResolver implement to resolve requests for the Rule's permissions field.
type RulePermissionsFieldResolver interface {
	// Permissions implements response to request for permissions field.
	Permissions(p graphql.ResolveParams) (interface{}, error)
}

//
// RuleFieldResolvers represents a collection of methods whose products represent the
// response values of the 'Rule' type.
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
type RuleFieldResolvers interface {
	RuleNamespaceFieldResolver
	RuleTypeFieldResolver
	RulePermissionsFieldResolver
}

// RuleAliases implements all methods on RuleFieldResolvers interface by using reflection to
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
type RuleAliases struct{}

// Namespace implements response to request for 'namespace' field.
func (_ RuleAliases) Namespace(p graphql.ResolveParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// Type implements response to request for 'type' field.
func (_ RuleAliases) Type(p graphql.ResolveParams) (RuleResource, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret := RuleResource(val.(string))
	return ret, err
}

// Permissions implements response to request for 'permissions' field.
func (_ RuleAliases) Permissions(p graphql.ResolveParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// RuleType Rule maps permissions to a given type
var RuleType = graphql.NewType("Rule", graphql.ObjectKind)

// RegisterRule registers Rule object type with given service.
func RegisterRule(svc *graphql.Service, impl RuleFieldResolvers) {
	svc.RegisterObject(_ObjectTypeRuleDesc, impl)
}
func _ObjTypeRuleNamespaceHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(RuleNamespaceFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Namespace(frp)
	}
}

func _ObjTypeRuleTypeHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(RuleTypeFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {

		val, err := resolver.Type(frp)
		return string(val), err
	}
}

func _ObjTypeRulePermissionsHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(RulePermissionsFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Permissions(frp)
	}
}

func _ObjectTypeRuleConfigFn() graphql1.ObjectConfig {
	return graphql1.ObjectConfig{
		Description: "Rule maps permissions to a given type",
		Fields: graphql1.Fields{
			"namespace": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "namespace in which this record resides",
				Name:              "namespace",
				Type:              graphql1.NewNonNull(graphql.OutputType("Namespace")),
			},
			"permissions": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "self descriptive",
				Name:              "permissions",
				Type:              graphql1.NewNonNull(graphql1.NewList(graphql1.NewNonNull(graphql.OutputType("RulePermission")))),
			},
			"type": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "resource the permissions apply to",
				Name:              "type",
				Type:              graphql1.NewNonNull(graphql.OutputType("RuleResource")),
			},
		},
		Interfaces: []*graphql1.Interface{},
		IsTypeOf: func(_ graphql1.IsTypeOfParams) bool {
			// NOTE:
			// Panic by default. Intent is that when Service is invoked, values of
			// these fields are updated with instantiated resolvers. If these
			// defaults are called it is most certainly programmer err.
			// If you're see this comment then: 'Whoops! Sorry, my bad.'
			panic("Unimplemented; see RuleFieldResolvers.")
		},
		Name: "Rule",
	}
}

// describe Rule's configuration; kept private to avoid unintentional tampering of configuration at runtime.
var _ObjectTypeRuleDesc = graphql.ObjectDesc{
	Config: _ObjectTypeRuleConfigFn,
	FieldHandlers: map[string]graphql.FieldHandler{
		"namespace":   _ObjTypeRuleNamespaceHandler,
		"permissions": _ObjTypeRulePermissionsHandler,
		"type":        _ObjTypeRuleTypeHandler,
	},
}

// RoleIDFieldResolver implement to resolve requests for the Role's id field.
type RoleIDFieldResolver interface {
	// ID implements response to request for id field.
	ID(p graphql.ResolveParams) (interface{}, error)
}

// RoleNameFieldResolver implement to resolve requests for the Role's name field.
type RoleNameFieldResolver interface {
	// Name implements response to request for name field.
	Name(p graphql.ResolveParams) (string, error)
}

// RoleRulesFieldResolver implement to resolve requests for the Role's rules field.
type RoleRulesFieldResolver interface {
	// Rules implements response to request for rules field.
	Rules(p graphql.ResolveParams) (interface{}, error)
}

//
// RoleFieldResolvers represents a collection of methods whose products represent the
// response values of the 'Role' type.
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
type RoleFieldResolvers interface {
	RoleIDFieldResolver
	RoleNameFieldResolver
	RoleRulesFieldResolver
}

// RoleAliases implements all methods on RoleFieldResolvers interface by using reflection to
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
type RoleAliases struct{}

// ID implements response to request for 'id' field.
func (_ RoleAliases) ID(p graphql.ResolveParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// Name implements response to request for 'name' field.
func (_ RoleAliases) Name(p graphql.ResolveParams) (string, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret := fmt.Sprint(val)
	return ret, err
}

// Rules implements response to request for 'rules' field.
func (_ RoleAliases) Rules(p graphql.ResolveParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// RoleType Role describes set of rules
var RoleType = graphql.NewType("Role", graphql.ObjectKind)

// RegisterRole registers Role object type with given service.
func RegisterRole(svc *graphql.Service, impl RoleFieldResolvers) {
	svc.RegisterObject(_ObjectTypeRoleDesc, impl)
}
func _ObjTypeRoleIDHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(RoleIDFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.ID(frp)
	}
}

func _ObjTypeRoleNameHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(RoleNameFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Name(frp)
	}
}

func _ObjTypeRoleRulesHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(RoleRulesFieldResolver)
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Rules(frp)
	}
}

func _ObjectTypeRoleConfigFn() graphql1.ObjectConfig {
	return graphql1.ObjectConfig{
		Description: "Role describes set of rules",
		Fields: graphql1.Fields{
			"id": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "self descriptive",
				Name:              "id",
				Type:              graphql1.NewNonNull(graphql1.ID),
			},
			"name": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "self descriptive",
				Name:              "name",
				Type:              graphql1.NewNonNull(graphql1.String),
			},
			"rules": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "self descriptive",
				Name:              "rules",
				Type:              graphql1.NewNonNull(graphql1.NewList(graphql1.NewNonNull(graphql.OutputType("Rule")))),
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
			panic("Unimplemented; see RoleFieldResolvers.")
		},
		Name: "Role",
	}
}

// describe Role's configuration; kept private to avoid unintentional tampering of configuration at runtime.
var _ObjectTypeRoleDesc = graphql.ObjectDesc{
	Config: _ObjectTypeRoleConfigFn,
	FieldHandlers: map[string]graphql.FieldHandler{
		"id":    _ObjTypeRoleIDHandler,
		"name":  _ObjTypeRoleNameHandler,
		"rules": _ObjTypeRoleRulesHandler,
	},
}

// RuleResource self descriptive
type RuleResource string

// RuleResources holds enum values
var RuleResources = _EnumTypeRuleResourceValues{
	ALL:           "ALL",
	ASSETS:        "ASSETS",
	CHECKS:        "CHECKS",
	ENTITIES:      "ENTITIES",
	HANDLERS:      "HANDLERS",
	HOOKS:         "HOOKS",
	MUTATORS:      "MUTATORS",
	ORGANIZATIONS: "ORGANIZATIONS",
	ROLES:         "ROLES",
	SILENCED:      "SILENCED",
	USERS:         "USERS",
}

// RuleResourceType self descriptive
var RuleResourceType = graphql.NewType("RuleResource", graphql.EnumKind)

// RegisterRuleResource registers RuleResource object type with given service.
func RegisterRuleResource(svc *graphql.Service) {
	svc.RegisterEnum(_EnumTypeRuleResourceDesc)
}
func _EnumTypeRuleResourceConfigFn() graphql1.EnumConfig {
	return graphql1.EnumConfig{
		Description: "self descriptive",
		Name:        "RuleResource",
		Values: graphql1.EnumValueConfigMap{
			"ALL": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "ALL",
			},
			"ASSETS": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "ASSETS",
			},
			"CHECKS": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "CHECKS",
			},
			"ENTITIES": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "ENTITIES",
			},
			"HANDLERS": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "HANDLERS",
			},
			"HOOKS": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "HOOKS",
			},
			"MUTATORS": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "MUTATORS",
			},
			"ORGANIZATIONS": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "ORGANIZATIONS",
			},
			"ROLES": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "ROLES",
			},
			"SILENCED": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "SILENCED",
			},
			"USERS": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "USERS",
			},
		},
	}
}

// describe RuleResource's configuration; kept private to avoid unintentional tampering of configuration at runtime.
var _EnumTypeRuleResourceDesc = graphql.EnumDesc{Config: _EnumTypeRuleResourceConfigFn}

type _EnumTypeRuleResourceValues struct {
	// ALL - self descriptive
	ALL RuleResource
	// ASSETS - self descriptive
	ASSETS RuleResource
	// CHECKS - self descriptive
	CHECKS RuleResource
	// ENTITIES - self descriptive
	ENTITIES RuleResource
	// HANDLERS - self descriptive
	HANDLERS RuleResource
	// HOOKS - self descriptive
	HOOKS RuleResource
	// MUTATORS - self descriptive
	MUTATORS RuleResource
	// ORGANIZATIONS - self descriptive
	ORGANIZATIONS RuleResource
	// ROLES - self descriptive
	ROLES RuleResource
	// SILENCED - self descriptive
	SILENCED RuleResource
	// USERS - self descriptive
	USERS RuleResource
}

// RulePermission self descriptive
type RulePermission string

// RulePermissions holds enum values
var RulePermissions = _EnumTypeRulePermissionValues{
	ALL:    "ALL",
	CREATE: "CREATE",
	DELETE: "DELETE",
	READ:   "READ",
	UPDATE: "UPDATE",
}

// RulePermissionType self descriptive
var RulePermissionType = graphql.NewType("RulePermission", graphql.EnumKind)

// RegisterRulePermission registers RulePermission object type with given service.
func RegisterRulePermission(svc *graphql.Service) {
	svc.RegisterEnum(_EnumTypeRulePermissionDesc)
}
func _EnumTypeRulePermissionConfigFn() graphql1.EnumConfig {
	return graphql1.EnumConfig{
		Description: "self descriptive",
		Name:        "RulePermission",
		Values: graphql1.EnumValueConfigMap{
			"ALL": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "ALL",
			},
			"CREATE": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "CREATE",
			},
			"DELETE": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "DELETE",
			},
			"READ": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "READ",
			},
			"UPDATE": &graphql1.EnumValueConfig{
				DeprecationReason: "",
				Description:       "self descriptive",
				Value:             "UPDATE",
			},
		},
	}
}

// describe RulePermission's configuration; kept private to avoid unintentional tampering of configuration at runtime.
var _EnumTypeRulePermissionDesc = graphql.EnumDesc{Config: _EnumTypeRulePermissionConfigFn}

type _EnumTypeRulePermissionValues struct {
	// ALL - self descriptive
	ALL RulePermission
	// CREATE - self descriptive
	CREATE RulePermission
	// READ - self descriptive
	READ RulePermission
	// UPDATE - self descriptive
	UPDATE RulePermission
	// DELETE - self descriptive
	DELETE RulePermission
}
