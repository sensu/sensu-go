package generator

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"github.com/graphql-go/graphql/language/ast"
)

func getNodeName(def ast.Node) string {
	switch d := def.(type) {
	case *ast.EnumDefinition:
		return d.Name.Value
	case *ast.InputObjectDefinition:
		return d.Name.Value
	case *ast.InterfaceDefinition:
		return d.Name.Value
	case *ast.ObjectDefinition:
		return d.Name.Value
	case *ast.ScalarDefinition:
		return d.Name.Value
	case *ast.UnionDefinition:
		return d.Name.Value
	default:
		return ""
	}
}

func getTypeStr(node ast.Node) string {
	switch node.(type) {
	case *ast.ScalarDefinition:
		return "Scalar"
	case *ast.ObjectDefinition:
		return "Object"
	case *ast.InputObjectDefinition:
		return "Input"
	case *ast.InterfaceDefinition:
		return "Interface"
	case *ast.UnionDefinition:
		return "Union"
	case *ast.EnumDefinition:
		return "Enum"
	default:
		fmt.Printf("node: %#v", node)
		panic("unknown node")
	}
}

func mkPrivatePrefix(node ast.Node) string {
	name := getNodeName(node)
	typeStr := getTypeStr(node)

	return fmt.Sprintf("_%sType%s", typeStr, name)
}

func mkPrivateID(node ast.Node, id string) string {
	return mkPrivatePrefix(node) + id
}

//
// Generate public func to register type with service
//
// == Example output
//
//   // RegisterDog registers Dog type with given service.
//   func RegisterDog(svc graphql.Service, impl DogFieldResolvers) {
//     svc.RegisterObject(_ObjTypeDogDesc, impl)
//   }
//
func genRegisterFn(node ast.Node, resolverImpl jen.Code) jen.Code {
	name := getNodeName(node)
	typeStr := getTypeStr(node)
	registerFnName := "Register" + name
	privateConfigName := mkPrivateID(node, "Desc")

	code := newGroup()
	code.Commentf(
		"%s registers %s object type with given service.",
		registerFnName,
		name,
	)
	code.
		Func().Id(registerFnName).
		ParamsFunc(func(g *jen.Group) {
			g.Id("svc").Op("*").Qual(servicePkg, "Service")
			if resolverImpl != nil {
				g.Id("impl").Add(resolverImpl)
			}
		}).
		Block(
			jen.
				Id("svc.Register" + typeStr).
				CallFunc(func(g *jen.Group) {
					g.Id(privateConfigName)
					if resolverImpl != nil {
						g.Id("impl")
					}
				}),
		)

	return code
}
