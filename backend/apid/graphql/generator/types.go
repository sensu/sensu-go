package generator

import (
	"github.com/dave/jennifer/jen"
	"github.com/jamesdphillips/graphql/language/ast"
)

func genMockInterfaceReference(t *ast.Named) *jen.Statement {
	return jen.Qual(utilPkg, "Interface").Call(jen.Lit(t.Name.Value))
}

func genMockObjectReference(t *ast.Named) *jen.Statement {
	return jen.Qual(utilPkg, "Object").Call(jen.Lit(t.Name.Value))
}

func genInputTypeReference(t ast.Type) *jen.Statement {
	return genTypeReference(t, "InputType")
}

func genOutputTypeReference(t ast.Type) *jen.Statement {
	return genTypeReference(t, "OutputType")
}

func genTypeReference(t ast.Type, expectedType string) *jen.Statement {
	var wrapperType ast.Type
	var namedType *ast.Named
	switch ttype := t.(type) {
	case *ast.List:
		wrapperType = ttype
		namedType = ttype.Type.(*ast.Named)
	case *ast.NonNull:
		wrapperType = t
		namedType = ttype.Type.(*ast.Named)
	case *ast.Named:
		namedType = ttype
	default:
		panic("unknown ast.Type given")
	}

	var valueStatement *jen.Statement
	switch namedType.Name.Value {
	case "Int":
		valueStatement = jen.Qual(graphqlPkg, "Int")
	case "Float":
		valueStatement = jen.Qual(graphqlPkg, "Float")
	case "String":
		valueStatement = jen.Qual(graphqlPkg, "String")
	case "Boolean":
		valueStatement = jen.Qual(graphqlPkg, "Boolean")
	case "DateTime":
		valueStatement = jen.Qual(graphqlPkg, "DateTime")
	default:
		name := namedType.Name.Value
		valueStatement = jen.Qual(utilPkg, expectedType).Call(jen.Lit(name))
	}

	if _, ok := wrapperType.(*ast.NonNull); ok {
		return jen.Qual(graphqlPkg, "NewNonNull").Call(valueStatement)
	} else if _, ok := wrapperType.(*ast.List); ok {
		return jen.Qual(graphqlPkg, "NewList").Call(valueStatement)
	}
	return valueStatement
}
