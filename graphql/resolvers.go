package graphql

import (
	"reflect"
	"strings"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
)

//
// ScalarResolver represents a collection of methods whose products represent
// the input and response values of a scalar type.
//
// == Example input SDL
//
//   """
//   Timestamps are great.
//   """
//   scalar Timestamp
//
// == Example implementation
//
//   // MyTimestampResolver implements ScalarResolver interface
//   type MyTimestampResolver struct {
//     defaultTZ *time.Location
//     logger    logrus.LogEntry
//   }
//
//   // Serialize serializes given date into RFC 943 compatible string.
//   func (r *MyTimestampResolver) Serialize(val interface{}) interface{} {
//     // ... implementation details ...
//   }
//
//   // ParseValue takes given value and coerces it into an instance of Time.
//   func (r *MyTimestampResolver) ParseValue(val interface{}) interface{} {
//     // ... implementation details ...
//     // eg. if val is an int use time.At(), if string time.Parse(), etc.
//   }
//
//   // ParseValue takes given value and coerces it into an instance of Time.
//   func (r *MyTimestampResolver) ParseValue(val ast.Value) interface{} {
//     // ... implementation details ...
//     //
//     // eg.
//     //
//     // if string value return value,
//     // if IntValue Atoi and return value,
//     // etc.
//   }`
//
type ScalarResolver interface {
	// Serialize an internal value to include in a response.
	Serialize(interface{}) interface{}

	// ParseValue parses an externally provided value to use as an input.
	ParseValue(interface{}) interface{}

	// ParseLiteral parses an externally provided literal value to use as an input.
	ParseLiteral(ast.Value) interface{}
}

//
// InterfaceTypeResolver represents a collection of methods whose products
// represent the input and response values of a interface type.
//
// == Example input SDL
//
//   "Pets are the bestest family members"
//   interface Pet {
//     "name of this fine beast."
//     name: String!
//   }
//
// == Example implementation
//
//   // PetResolver implements InterfaceTypeResolver
//   type PetResolver struct {
//     logger    logrus.LogEntry
//   }
//
//   // ResolveType should return type reference
//   func (r *PetResolver) ResolveType(val interface {}, _ graphql.ResolveTypeParams) graphql.Type {
//     // ... implementation details ...
//     switch pet := val.(type) {
//     when *Dog:
//       return schema.DogType // Handled by type identified by 'Dog'
//     when *Cat:
//       return schema.CatType // Handled by type identified by 'Cat'
//     }
//     panic("Unimplemented")
//   }`,
//
type InterfaceTypeResolver interface {
	ResolveType(interface{}, ResolveTypeParams) *Type
}

//
// UnionTypeResolver represents a collection of methods whose products
// represent the input and response values of a union type.
//
// == Example input SDL
//
//   """
//   Feed includes all stuff and things.
//   """
//   union Feed = Story | Article | Advert
//
// == Example implementation
//
//   // FeedResolver implements UnionTypeResolver
//   type FeedResolver struct {
//     logger logrus.LogEntry
//   }
//
//   // ResolveType should return type reference
//   func (r *FeedResolver) ResolveType(val interface {}, _ graphql.ResolveTypeParams) graphql.Type {
//     // ... implementation details ...
//     switch entity := val.(type) {
//     when *Article:
//       return schema.ArticleType
//     when *Story:
//       return schema.StoreType
//     when *Advert:
//       return schema.AdvertType
//     }
//     panic("Unimplemented")
//   }
//
type UnionTypeResolver interface {
	ResolveType(interface{}, ResolveTypeParams) *Type
}

// DefaultResolver uses reflection to attempt to resolve the result of a given
// field.
//
// Heavily borrows from: https://github.com/graphql-go/graphql/blob/9b68c99d07d901738c15564ec1a0f57d07d884a7/executor.go#L823-L881
func DefaultResolver(source interface{}, fieldName string) (interface{}, error) {
	sourceVal := reflect.ValueOf(source)
	if sourceVal.IsValid() && sourceVal.Type().Kind() == reflect.Ptr {
		sourceVal = sourceVal.Elem()
	}
	if !sourceVal.IsValid() {
		return nil, nil
	}

	// Struct
	if sourceVal.Type().Kind() == reflect.Struct {
		fieldName = strings.Title(fieldName)
		for i := 0; i < sourceVal.NumField(); i++ {
			valueField := sourceVal.Field(i)
			typeField := sourceVal.Type().Field(i)
			if typeField.Name == fieldName {
				// If ptr and value is nil return nil
				if valueField.Type().Kind() == reflect.Ptr && valueField.IsNil() {
					return nil, nil
				}
				return valueField.Interface(), nil
			}
			tag := typeField.Tag
			checkTag := func(tagName string) bool {
				t := tag.Get(tagName)
				tOptions := strings.Split(t, ",")
				if len(tOptions) == 0 {
					return false
				}
				if tOptions[0] != fieldName {
					return false
				}
				return true
			}
			if checkTag("json") || checkTag("graphql") {
				return valueField.Interface(), nil
			}
			if valueField.Kind() == reflect.Struct && typeField.Anonymous {
				return DefaultResolver(valueField.Interface(), fieldName)
			}
			continue
		}
		return nil, nil
	}

	// map[string]interface
	if sourceMap, ok := source.(map[string]interface{}); ok {
		property := sourceMap[fieldName]
		val := reflect.ValueOf(property)
		if val.IsValid() && val.Type().Kind() == reflect.Func {
			// try type casting the func to the most basic func signature
			// for more complex signatures, user have to define ResolveFn
			if propertyFn, ok := property.(func() interface{}); ok {
				return propertyFn(), nil
			}
		}
		return property, nil
	}

	// last resort, return nil
	return nil, nil
}

type typeResolver interface {
	ResolveType(interface{}, ResolveTypeParams) *Type
}

func newResolveTypeFn(typeMap graphql.TypeMap, impl interface{}) graphql.ResolveTypeFn {
	resolver := impl.(typeResolver)
	return func(p graphql.ResolveTypeParams) *graphql.Object {
		typeRef := resolver.ResolveType(p.Value, p)
		if typeRef == nil {
			return nil
		}

		objType, ok := typeMap[typeRef.Name()]
		if !ok {
			return nil
		}

		return objType.(*graphql.Object)
	}
}

type isTypeOfResolver interface {
	IsTypeOf(interface{}, IsTypeOfParams) bool
}

func newIsTypeOfFn(impl interface{}) graphql.IsTypeOfFn {
	resolver := impl.(isTypeOfResolver)
	return func(p graphql.IsTypeOfParams) bool {
		return resolver.IsTypeOf(p.Value, p)
	}
}
