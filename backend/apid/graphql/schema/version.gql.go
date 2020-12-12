// Code generated by scripts/gengraphql.go. DO NOT EDIT.

package schema

import (
	errors "errors"
	graphql1 "github.com/graphql-go/graphql"
	graphql "github.com/sensu/sensu-go/graphql"
	time "time"
)

//
// VersionsFieldResolvers represents a collection of methods whose products represent the
// response values of the 'Versions' type.
type VersionsFieldResolvers interface {
	// Etcd implements response to request for 'etcd' field.
	Etcd(p graphql.ResolveParams) (interface{}, error)

	// Backend implements response to request for 'backend' field.
	Backend(p graphql.ResolveParams) (interface{}, error)
}

// VersionsAliases implements all methods on VersionsFieldResolvers interface by using reflection to
// match name of field to a field on the given value. Intent is reduce friction
// of writing new resolvers by removing all the instances where you would simply
// have the resolvers method return a field.
type VersionsAliases struct{}

// Etcd implements response to request for 'etcd' field.
func (_ VersionsAliases) Etcd(p graphql.ResolveParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// Backend implements response to request for 'backend' field.
func (_ VersionsAliases) Backend(p graphql.ResolveParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// VersionsType Describes the version of different components of the system.
var VersionsType = graphql.NewType("Versions", graphql.ObjectKind)

// RegisterVersions registers Versions object type with given service.
func RegisterVersions(svc *graphql.Service, impl VersionsFieldResolvers) {
	svc.RegisterObject(_ObjectTypeVersionsDesc, impl)
}
func _ObjTypeVersionsEtcdHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(interface {
		Etcd(p graphql.ResolveParams) (interface{}, error)
	})
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Etcd(frp)
	}
}

func _ObjTypeVersionsBackendHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(interface {
		Backend(p graphql.ResolveParams) (interface{}, error)
	})
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Backend(frp)
	}
}

func _ObjectTypeVersionsConfigFn() graphql1.ObjectConfig {
	return graphql1.ObjectConfig{
		Description: "Describes the version of different components of the system.",
		Fields: graphql1.Fields{
			"backend": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "self descriptive",
				Name:              "backend",
				Type:              graphql.OutputType("SensuBackendVersion"),
			},
			"etcd": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "self descriptive",
				Name:              "etcd",
				Type:              graphql.OutputType("EtcdVersions"),
			},
		},
		Interfaces: []*graphql1.Interface{},
		IsTypeOf: func(_ graphql1.IsTypeOfParams) bool {
			// NOTE:
			// Panic by default. Intent is that when Service is invoked, values of
			// these fields are updated with instantiated resolvers. If these
			// defaults are called it is most certainly programmer err.
			// If you're see this comment then: 'Whoops! Sorry, my bad.'
			panic("Unimplemented; see VersionsFieldResolvers.")
		},
		Name: "Versions",
	}
}

// describe Versions's configuration; kept private to avoid unintentional tampering of configuration at runtime.
var _ObjectTypeVersionsDesc = graphql.ObjectDesc{
	Config: _ObjectTypeVersionsConfigFn,
	FieldHandlers: map[string]graphql.FieldHandler{
		"backend": _ObjTypeVersionsBackendHandler,
		"etcd":    _ObjTypeVersionsEtcdHandler,
	},
}

//
// EtcdVersionsFieldResolvers represents a collection of methods whose products represent the
// response values of the 'EtcdVersions' type.
type EtcdVersionsFieldResolvers interface {
	// Server implements response to request for 'server' field.
	Server(p graphql.ResolveParams) (string, error)

	// Cluster implements response to request for 'cluster' field.
	Cluster(p graphql.ResolveParams) (string, error)
}

// EtcdVersionsAliases implements all methods on EtcdVersionsFieldResolvers interface by using reflection to
// match name of field to a field on the given value. Intent is reduce friction
// of writing new resolvers by removing all the instances where you would simply
// have the resolvers method return a field.
type EtcdVersionsAliases struct{}

// Server implements response to request for 'server' field.
func (_ EtcdVersionsAliases) Server(p graphql.ResolveParams) (string, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret, ok := val.(string)
	if err != nil {
		return ret, err
	}
	if !ok {
		return ret, errors.New("unable to coerce value for field 'server'")
	}
	return ret, err
}

// Cluster implements response to request for 'cluster' field.
func (_ EtcdVersionsAliases) Cluster(p graphql.ResolveParams) (string, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret, ok := val.(string)
	if err != nil {
		return ret, err
	}
	if !ok {
		return ret, errors.New("unable to coerce value for field 'cluster'")
	}
	return ret, err
}

// EtcdVersionsType Describes the version of Etcd instance and the Etcd cluster.
var EtcdVersionsType = graphql.NewType("EtcdVersions", graphql.ObjectKind)

// RegisterEtcdVersions registers EtcdVersions object type with given service.
func RegisterEtcdVersions(svc *graphql.Service, impl EtcdVersionsFieldResolvers) {
	svc.RegisterObject(_ObjectTypeEtcdVersionsDesc, impl)
}
func _ObjTypeEtcdVersionsServerHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(interface {
		Server(p graphql.ResolveParams) (string, error)
	})
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Server(frp)
	}
}

func _ObjTypeEtcdVersionsClusterHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(interface {
		Cluster(p graphql.ResolveParams) (string, error)
	})
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Cluster(frp)
	}
}

func _ObjectTypeEtcdVersionsConfigFn() graphql1.ObjectConfig {
	return graphql1.ObjectConfig{
		Description: "Describes the version of Etcd instance and the Etcd cluster.",
		Fields: graphql1.Fields{
			"cluster": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "Etcd cluster version; adheres to semver.",
				Name:              "cluster",
				Type:              graphql1.NewNonNull(graphql1.String),
			},
			"server": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "Etcd version; adheres to semver.",
				Name:              "server",
				Type:              graphql1.NewNonNull(graphql1.String),
			},
		},
		Interfaces: []*graphql1.Interface{},
		IsTypeOf: func(_ graphql1.IsTypeOfParams) bool {
			// NOTE:
			// Panic by default. Intent is that when Service is invoked, values of
			// these fields are updated with instantiated resolvers. If these
			// defaults are called it is most certainly programmer err.
			// If you're see this comment then: 'Whoops! Sorry, my bad.'
			panic("Unimplemented; see EtcdVersionsFieldResolvers.")
		},
		Name: "EtcdVersions",
	}
}

// describe EtcdVersions's configuration; kept private to avoid unintentional tampering of configuration at runtime.
var _ObjectTypeEtcdVersionsDesc = graphql.ObjectDesc{
	Config: _ObjectTypeEtcdVersionsConfigFn,
	FieldHandlers: map[string]graphql.FieldHandler{
		"cluster": _ObjTypeEtcdVersionsClusterHandler,
		"server":  _ObjTypeEtcdVersionsServerHandler,
	},
}

//
// SensuBackendVersionFieldResolvers represents a collection of methods whose products represent the
// response values of the 'SensuBackendVersion' type.
type SensuBackendVersionFieldResolvers interface {
	// Version implements response to request for 'version' field.
	Version(p graphql.ResolveParams) (string, error)

	// BuildSHA implements response to request for 'buildSHA' field.
	BuildSHA(p graphql.ResolveParams) (string, error)

	// BuildDate implements response to request for 'buildDate' field.
	BuildDate(p graphql.ResolveParams) (*time.Time, error)
}

// SensuBackendVersionAliases implements all methods on SensuBackendVersionFieldResolvers interface by using reflection to
// match name of field to a field on the given value. Intent is reduce friction
// of writing new resolvers by removing all the instances where you would simply
// have the resolvers method return a field.
type SensuBackendVersionAliases struct{}

// Version implements response to request for 'version' field.
func (_ SensuBackendVersionAliases) Version(p graphql.ResolveParams) (string, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret, ok := val.(string)
	if err != nil {
		return ret, err
	}
	if !ok {
		return ret, errors.New("unable to coerce value for field 'version'")
	}
	return ret, err
}

// BuildSHA implements response to request for 'buildSHA' field.
func (_ SensuBackendVersionAliases) BuildSHA(p graphql.ResolveParams) (string, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret, ok := val.(string)
	if err != nil {
		return ret, err
	}
	if !ok {
		return ret, errors.New("unable to coerce value for field 'buildSHA'")
	}
	return ret, err
}

// BuildDate implements response to request for 'buildDate' field.
func (_ SensuBackendVersionAliases) BuildDate(p graphql.ResolveParams) (*time.Time, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret, ok := val.(*time.Time)
	if err != nil {
		return ret, err
	}
	if !ok {
		return ret, errors.New("unable to coerce value for field 'buildDate'")
	}
	return ret, err
}

// SensuBackendVersionType Describes the version of the Sensu backend node.
var SensuBackendVersionType = graphql.NewType("SensuBackendVersion", graphql.ObjectKind)

// RegisterSensuBackendVersion registers SensuBackendVersion object type with given service.
func RegisterSensuBackendVersion(svc *graphql.Service, impl SensuBackendVersionFieldResolvers) {
	svc.RegisterObject(_ObjectTypeSensuBackendVersionDesc, impl)
}
func _ObjTypeSensuBackendVersionVersionHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(interface {
		Version(p graphql.ResolveParams) (string, error)
	})
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Version(frp)
	}
}

func _ObjTypeSensuBackendVersionBuildSHAHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(interface {
		BuildSHA(p graphql.ResolveParams) (string, error)
	})
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.BuildSHA(frp)
	}
}

func _ObjTypeSensuBackendVersionBuildDateHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(interface {
		BuildDate(p graphql.ResolveParams) (*time.Time, error)
	})
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.BuildDate(frp)
	}
}

func _ObjectTypeSensuBackendVersionConfigFn() graphql1.ObjectConfig {
	return graphql1.ObjectConfig{
		Description: "Describes the version of the Sensu backend node.",
		Fields: graphql1.Fields{
			"buildDate": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "self descriptive",
				Name:              "buildDate",
				Type:              graphql1.DateTime,
			},
			"buildSHA": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "self descriptive",
				Name:              "buildSHA",
				Type:              graphql1.String,
			},
			"version": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "Version of the current node; adheres to semver.",
				Name:              "version",
				Type:              graphql1.String,
			},
		},
		Interfaces: []*graphql1.Interface{},
		IsTypeOf: func(_ graphql1.IsTypeOfParams) bool {
			// NOTE:
			// Panic by default. Intent is that when Service is invoked, values of
			// these fields are updated with instantiated resolvers. If these
			// defaults are called it is most certainly programmer err.
			// If you're see this comment then: 'Whoops! Sorry, my bad.'
			panic("Unimplemented; see SensuBackendVersionFieldResolvers.")
		},
		Name: "SensuBackendVersion",
	}
}

// describe SensuBackendVersion's configuration; kept private to avoid unintentional tampering of configuration at runtime.
var _ObjectTypeSensuBackendVersionDesc = graphql.ObjectDesc{
	Config: _ObjectTypeSensuBackendVersionConfigFn,
	FieldHandlers: map[string]graphql.FieldHandler{
		"buildDate": _ObjTypeSensuBackendVersionBuildDateHandler,
		"buildSHA":  _ObjTypeSensuBackendVersionBuildSHAHandler,
		"version":   _ObjTypeSensuBackendVersionVersionHandler,
	},
}
