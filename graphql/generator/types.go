package generator

import (
	"github.com/dave/jennifer/jen"
	"github.com/jamesdphillips/graphql/language/ast"
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
		return jen.Qual(graphqlGoPkg, "NewList").Call(s)
	case *ast.NonNull:
		s := genTypeReference(ttype.Type, expectedType)
		return jen.Qual(graphqlGoPkg, "NewNonNull").Call(s)
	case *ast.Named:
		namedType = ttype
	default:
		panic("unknown ast.Type given")
	}

	var valueStatement *jen.Statement
	switch namedType.Name.Value {
	case "Int":
		valueStatement = jen.Qual(graphqlGoPkg, "Int")
	case "Float":
		valueStatement = jen.Qual(graphqlGoPkg, "Float")
	case "String":
		valueStatement = jen.Qual(graphqlGoPkg, "String")
	case "Boolean":
		valueStatement = jen.Qual(graphqlGoPkg, "Boolean")
	case "DateTime":
		valueStatement = jen.Qual(graphqlGoPkg, "DateTime")
	default:
		name := namedType.Name.Value
		valueStatement = jen.Qual(servicePkg, expectedType).Call(jen.Lit(name))
	}

	return valueStatement
}
