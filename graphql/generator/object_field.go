package generator

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"github.com/graphql-go/graphql/language/ast"
)

func genFieldResolversInterface(name string, node *ast.ObjectDefinition, i info) jen.Code {
	interfaceName := mkFieldResolversName(name)

	code := newGroup()
	code.Commentf(`//
// %s represents a collection of methods whose products represent the
// response values of the '%s' type.`,
		interfaceName,
		name,
	)
	// Generate resolver interface.
	code.
		Type().Id(interfaceName).
		InterfaceFunc(func(g *jen.Group) {
			// Include each field resolver
			for _, field := range node.Fields {
				g.Commentf(
					"%s implements response to request for '%s' field.",
					toFieldName(field.Name.Value),
					field.Name.Value,
				)
				g.Add(genFieldResolverSignature(field, i))
				g.Line()
			}
			g.Line()
		})
	return code
}

func genFieldAliases(name string, node *ast.ObjectDefinition, i info) jen.Code {
	fieldResolversName := mkFieldResolversName(name)
	aliasResolver := fmt.Sprintf("%sAliases", name)

	code := newGroup()
	code.Commentf(`// %s implements all methods on %s interface by using reflection to
// match name of field to a field on the given value. Intent is reduce friction
// of writing new resolvers by removing all the instances where you would simply
// have the resolvers method return a field.`,
		aliasResolver,
		fieldResolversName,
	)
	code.Type().Id(aliasResolver).Struct()
	for _, field := range node.Fields {
		// Define method for each field in object type
		name := field.Name.Value
		titleizedName := toFieldName(field.Name.Value)
		resolverFnSignature := genFieldResolverSignature(field, i)
		coerceType := genFieldResolverTypeCoercion(field.Type, i, true)

		code.Commentf("%s implements response to request for '%s' field.", titleizedName, name)
		code.
			Func().Params(jen.Id("_").Id(aliasResolver)).
			Add(resolverFnSignature).
			BlockFunc(func(g *jen.Group) {
				g.List(jen.List(jen.Id("val"), jen.Id("err"))).Op(":=").
					Qual(servicePkg, "DefaultResolver").
					Call(
						jen.Id("p").Dot("Source"),
						jen.Id("p").Dot("Info").Dot("FieldName"),
					)
				if coerceType != nil {
					g.List(jen.Id("ret"), jen.Id("ok")).Op(":=").Add(coerceType)
					g.If(jen.Id("err").Op("!=").Nil()).Block(
						jen.Return(jen.List(jen.Id("ret"), jen.Id("err"))),
					)
					g.If(jen.Op("!").Id("ok")).Block(
						jen.Return(jen.List(
							jen.Id("ret"),
							jen.Qual("errors", "New").Call(
								jen.Lit(fmt.Sprintf("unable to coerce value for field '%s'", name)),
							),
						)),
					)
					g.Return(jen.List(jen.Id("ret"), jen.Id("err")))
				} else {
					g.Return(jen.List(jen.Id("val"), jen.Id("err")))
				}
			})
	}
	return code
}

func mkFieldResolversName(prefix string) string {
	return prefix + "FieldResolvers"
}
