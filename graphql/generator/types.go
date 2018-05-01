package generator

import (
	"github.com/dave/jennifer/jen"
	"github.com/graphql-go/graphql/language/ast"
)

func genMockInterfaceReference(t *ast.Named) *jen.Statement {
	return jen.Qual(servicePkg, "Interface").Call(jen.Lit(t.Name.Value))
}

func genMockObjectReference(t *ast.Named) *jen.Statement {
	return jen.Qual(servicePkg, "Object").Call(jen.Lit(t.Name.Value))
}

func genInputTypeReference(t ast.Type) *jen.Statement {
	return genTypeReference(t, "InputType")
}

func genOutputTypeReference(t ast.Type) *jen.Statement {
	return genTypeReference(t, "OutputType")
}

func genTypeReference(t ast.Type, expectedType string) *jen.Statement {
	var namedType *ast.Named
	switch ttype := t.(type) {
	case *ast.List:
		s := genTypeReference(ttype.Type, expectedType)
		return jen.Qual(defsPkg, "NewList").Call(s)
	case *ast.NonNull:
		s := genTypeReference(ttype.Type, expectedType)
		return jen.Qual(defsPkg, "NewNonNull").Call(s)
	case *ast.Named:
		namedType = ttype
	default:
		panic("unknown ast.Type given")
	}

	var valueStatement *jen.Statement
	switch namedType.Name.Value {
	case "Int":
		valueStatement = jen.Qual(defsPkg, "Int")
	case "Float":
		valueStatement = jen.Qual(defsPkg, "Float")
	case "String":
		valueStatement = jen.Qual(defsPkg, "String")
	case "Boolean":
		valueStatement = jen.Qual(defsPkg, "Boolean")
	case "DateTime":
		valueStatement = jen.Qual(defsPkg, "DateTime")
	case "ID":
		valueStatement = jen.Qual(defsPkg, "ID")
	default:
		name := namedType.Name.Value
		valueStatement = jen.Qual(servicePkg, expectedType).Call(jen.Lit(name))
	}

	return valueStatement
}

func genConcreteTypeReference(t ast.Type, i info) jen.Code {
	var namedType *ast.Named
	switch ttype := t.(type) {
	case *ast.List:
		s := genConcreteTypeReference(ttype.Type, i)
		return jen.Index().Add(s)
	case *ast.NonNull:
		return genConcreteTypeReference(ttype.Type, i)
	case *ast.Named:
		namedType = ttype
	default:
		panic("unknown ast.Type given")
	}

	if code := genBuiltinTypeReference(namedType); code != nil {
		return code
	}

	typeName := namedType.Name.Value
	if matchedDef, ok := i.definitions[typeName]; ok {
		// if name matches an input object or enum definition use it.
		if _, ok := matchedDef.(*ast.InputObjectDefinition); ok {
			return jen.Op("*").Id(typeName)
		} else if _, ok := matchedDef.(*ast.EnumDefinition); ok {
			return jen.Id(typeName)
		}
	}
	return jen.Interface()
}

func genBuiltinTypeReference(t *ast.Named) jen.Code {
	switch t.Name.Value {
	//
	// The `ID` scalar type represents a unique identifier, often used to
	// refetch an object or as key for a cache. The ID type appears in a JSON
	// response as a String; however, it is not intended to be human-readable.
	// When expected as an input type, any string (such as `\"4\"`) or integer
	// (such as `4`) input value will be accepted as an ID.
	//
	// As such, when resolving a field we can safely assume the value encountered
	// will always be a string value.
	//
	// Spec: http://facebook.github.io/graphql/October2016/#sec-ID
	//
	case "ID":
		return jen.String()
	case "Int":
		return jen.Int()
	case "Float":
		return jen.Float64()
	case "String":
		return jen.String()
	case "Boolean":
		return jen.Bool()
	case "DateTime":
		return jen.Qual("time", "Time")
	}
	return nil
}
