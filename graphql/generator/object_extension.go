package generator

import (
	"github.com/dave/jennifer/jen"
	"github.com/graphql-go/graphql/language/ast"
)

const (
	MissingNamedDirectiveErr = `extend type must be followed by @named(suffix: "MyUniqueId") directive`
)

func genObjectExtension(node *ast.TypeExtensionDefinition, i info) jen.Code {
	code := newGroup()

	// Name
	suffix := mustExtractSuffix(node.Definition)
	tname := node.Definition.GetName().Value
	name := tname + "Extension" + suffix

	// Ids ...
	fieldResolverName := mkFieldResolversName(name)
	privateConfigName := mkPrivateID(node, "Desc")
	privateConfigThunkName := mkPrivateID(node, "ConfigFn")

	// Generate field resolver interfaces
	for _, f := range node.Definition.Fields {
		resolverCode := genFieldResolverInterface(f, i)
		code.Add(resolverCode)
		code.Line()
	}

	// Generate resolver interface
	code.Add(genFieldResolversInterface(name, node.Definition, i))
	code.Line()

	// Generate public func to register type with service
	code.Add(genRegisterFn(node, jen.Id(fieldResolverName)))
	code.Line()

	// Generate field handlers
	for _, f := range node.Definition.Fields {
		handler := genFieldHandlerFn(f, i)
		code.Add(handler)
		code.Line()
	}

	// Generate interface references
	ints := jen.
		Index().Op("*").Qual(defsPkg, "Interface").
		ValuesFunc(
			func(g *jen.Group) {
				for _, n := range node.Definition.Interfaces {
					g.Line().Add(genMockInterfaceReference(n))
				}
			},
		)

	// Generate config thunk
	code.
		Func().Id(privateConfigThunkName).
		Params().Qual(defsPkg, "ObjectConfig").
		Block(
			jen.Return(jen.Qual(defsPkg, "ObjectConfig").Values(jen.Dict{
				jen.Id("Name"):        jen.Lit(node.Definition.Name.Value),
				jen.Id("Description"): jen.Lit(""),
				jen.Id("Interfaces"):  ints,
				jen.Id("Fields"):      genFields(node.Definition.Fields),
			})),
		)

	// Generate type description
	code.Commentf(
		`describe %s's configuration; kept private to avoid unintentional tampering of configuration at runtime.`,
		name,
	)
	code.
		Var().Id(privateConfigName).Op("=").
		Qual(servicePkg, "ObjectDesc").
		Values(jen.Dict{
			jen.Id("Config"): jen.Id(privateConfigThunkName),
			jen.Id("FieldHandlers"): jen.Map(jen.String()).Qual(servicePkg, "FieldHandler").Values(jen.DictFunc(func(d jen.Dict) {
				for _, f := range node.Definition.Fields {
					key := f.Name.Value
					handlerName := genFieldHandlerName(f, i)
					d[jen.Lit(key)] = jen.Id(handlerName)
				}
			})),
		})

	return code
}
