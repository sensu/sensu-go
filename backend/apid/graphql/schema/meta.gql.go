// Code generated by scripts/gengraphql.go. DO NOT EDIT.

package schema

import (
	errors "errors"
	graphql1 "github.com/graphql-go/graphql"
	graphql "github.com/sensu/sensu-go/graphql"
)

// KVPairStringFieldResolvers represents a collection of methods whose products represent the
// response values of the 'KVPairString' type.
type KVPairStringFieldResolvers interface {
	// Key implements response to request for 'key' field.
	Key(p graphql.ResolveParams) (string, error)

	// Val implements response to request for 'val' field.
	Val(p graphql.ResolveParams) (string, error)
}

// KVPairStringAliases implements all methods on KVPairStringFieldResolvers interface by using reflection to
// match name of field to a field on the given value. Intent is reduce friction
// of writing new resolvers by removing all the instances where you would simply
// have the resolvers method return a field.
type KVPairStringAliases struct{}

// Key implements response to request for 'key' field.
func (_ KVPairStringAliases) Key(p graphql.ResolveParams) (string, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret, ok := val.(string)
	if err != nil {
		return ret, err
	}
	if !ok {
		return ret, errors.New("unable to coerce value for field 'key'")
	}
	return ret, err
}

// Val implements response to request for 'val' field.
func (_ KVPairStringAliases) Val(p graphql.ResolveParams) (string, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret, ok := val.(string)
	if err != nil {
		return ret, err
	}
	if !ok {
		return ret, errors.New("unable to coerce value for field 'val'")
	}
	return ret, err
}

/*
KVPairStringType The KVPairString type respresents a name-value relationship where the value is
always a string.
*/
var KVPairStringType = graphql.NewType("KVPairString", graphql.ObjectKind)

// RegisterKVPairString registers KVPairString object type with given service.
func RegisterKVPairString(svc *graphql.Service, impl KVPairStringFieldResolvers) {
	svc.RegisterObject(_ObjectTypeKVPairStringDesc, impl)
}
func _ObjTypeKVPairStringKeyHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(interface {
		Key(p graphql.ResolveParams) (string, error)
	})
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Key(frp)
	}
}

func _ObjTypeKVPairStringValHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(interface {
		Val(p graphql.ResolveParams) (string, error)
	})
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Val(frp)
	}
}

func _ObjectTypeKVPairStringConfigFn() graphql1.ObjectConfig {
	return graphql1.ObjectConfig{
		Description: "The KVPairString type respresents a name-value relationship where the value is\nalways a string.",
		Fields: graphql1.Fields{
			"key": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "self descriptive",
				Name:              "key",
				Type:              graphql1.NewNonNull(graphql1.String),
			},
			"val": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "self descriptive",
				Name:              "val",
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
			panic("Unimplemented; see KVPairStringFieldResolvers.")
		},
		Name: "KVPairString",
	}
}

// describe KVPairString's configuration; kept private to avoid unintentional tampering of configuration at runtime.
var _ObjectTypeKVPairStringDesc = graphql.ObjectDesc{
	Config: _ObjectTypeKVPairStringConfigFn,
	FieldHandlers: map[string]graphql.FieldHandler{
		"key": _ObjTypeKVPairStringKeyHandler,
		"val": _ObjTypeKVPairStringValHandler,
	},
}

// ObjectMetaFieldResolvers represents a collection of methods whose products represent the
// response values of the 'ObjectMeta' type.
type ObjectMetaFieldResolvers interface {
	// Name implements response to request for 'name' field.
	Name(p graphql.ResolveParams) (string, error)

	// Namespace implements response to request for 'namespace' field.
	Namespace(p graphql.ResolveParams) (string, error)

	// Labels implements response to request for 'labels' field.
	Labels(p graphql.ResolveParams) (interface{}, error)

	// Annotations implements response to request for 'annotations' field.
	Annotations(p graphql.ResolveParams) (interface{}, error)

	// CreatedBy implements response to request for 'createdBy' field.
	CreatedBy(p graphql.ResolveParams) (string, error)
}

// ObjectMetaAliases implements all methods on ObjectMetaFieldResolvers interface by using reflection to
// match name of field to a field on the given value. Intent is reduce friction
// of writing new resolvers by removing all the instances where you would simply
// have the resolvers method return a field.
type ObjectMetaAliases struct{}

// Name implements response to request for 'name' field.
func (_ ObjectMetaAliases) Name(p graphql.ResolveParams) (string, error) {
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

// Namespace implements response to request for 'namespace' field.
func (_ ObjectMetaAliases) Namespace(p graphql.ResolveParams) (string, error) {
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

// Labels implements response to request for 'labels' field.
func (_ ObjectMetaAliases) Labels(p graphql.ResolveParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// Annotations implements response to request for 'annotations' field.
func (_ ObjectMetaAliases) Annotations(p graphql.ResolveParams) (interface{}, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	return val, err
}

// CreatedBy implements response to request for 'createdBy' field.
func (_ ObjectMetaAliases) CreatedBy(p graphql.ResolveParams) (string, error) {
	val, err := graphql.DefaultResolver(p.Source, p.Info.FieldName)
	ret, ok := val.(string)
	if err != nil {
		return ret, err
	}
	if !ok {
		return ret, errors.New("unable to coerce value for field 'createdBy'")
	}
	return ret, err
}

// ObjectMetaType ObjectMeta is metadata all persisted objects have.
var ObjectMetaType = graphql.NewType("ObjectMeta", graphql.ObjectKind)

// RegisterObjectMeta registers ObjectMeta object type with given service.
func RegisterObjectMeta(svc *graphql.Service, impl ObjectMetaFieldResolvers) {
	svc.RegisterObject(_ObjectTypeObjectMetaDesc, impl)
}
func _ObjTypeObjectMetaNameHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(interface {
		Name(p graphql.ResolveParams) (string, error)
	})
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Name(frp)
	}
}

func _ObjTypeObjectMetaNamespaceHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(interface {
		Namespace(p graphql.ResolveParams) (string, error)
	})
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Namespace(frp)
	}
}

func _ObjTypeObjectMetaLabelsHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(interface {
		Labels(p graphql.ResolveParams) (interface{}, error)
	})
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Labels(frp)
	}
}

func _ObjTypeObjectMetaAnnotationsHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(interface {
		Annotations(p graphql.ResolveParams) (interface{}, error)
	})
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.Annotations(frp)
	}
}

func _ObjTypeObjectMetaCreatedByHandler(impl interface{}) graphql1.FieldResolveFn {
	resolver := impl.(interface {
		CreatedBy(p graphql.ResolveParams) (string, error)
	})
	return func(frp graphql1.ResolveParams) (interface{}, error) {
		return resolver.CreatedBy(frp)
	}
}

func _ObjectTypeObjectMetaConfigFn() graphql1.ObjectConfig {
	return graphql1.ObjectConfig{
		Description: "ObjectMeta is metadata all persisted objects have.",
		Fields: graphql1.Fields{
			"annotations": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "Annotations is an unstructured key value map stored with a resource that\nmay be set by external tools to store and retrieve arbitrary metadata. They\nare not queryable and should be preserved when modifying objects.",
				Name:              "annotations",
				Type:              graphql1.NewList(graphql1.NewNonNull(graphql.OutputType("KVPairString"))),
			},
			"createdBy": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "CreatedBy field indicates which user created the resource",
				Name:              "createdBy",
				Type:              graphql1.NewNonNull(graphql1.String),
			},
			"labels": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "Map of string keys and values that can be used to organize and categorize\n(scope and select) objects. May also be used in filters and token\nsubstitution.",
				Name:              "labels",
				Type:              graphql1.NewList(graphql1.NewNonNull(graphql.OutputType("KVPairString"))),
			},
			"name": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "Name must be unique within a namespace. Name is primarily intended for\ncreation idempotence and configuration definition.",
				Name:              "name",
				Type:              graphql1.NewNonNull(graphql1.String),
			},
			"namespace": &graphql1.Field{
				Args:              graphql1.FieldConfigArgument{},
				DeprecationReason: "",
				Description:       "Namespace defines a logical grouping of objects within which each object\nname must be unique.",
				Name:              "namespace",
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
			panic("Unimplemented; see ObjectMetaFieldResolvers.")
		},
		Name: "ObjectMeta",
	}
}

// describe ObjectMeta's configuration; kept private to avoid unintentional tampering of configuration at runtime.
var _ObjectTypeObjectMetaDesc = graphql.ObjectDesc{
	Config: _ObjectTypeObjectMetaConfigFn,
	FieldHandlers: map[string]graphql.FieldHandler{
		"annotations": _ObjTypeObjectMetaAnnotationsHandler,
		"createdBy":   _ObjTypeObjectMetaCreatedByHandler,
		"labels":      _ObjTypeObjectMetaLabelsHandler,
		"name":        _ObjTypeObjectMetaNameHandler,
		"namespace":   _ObjTypeObjectMetaNamespaceHandler,
	},
}
