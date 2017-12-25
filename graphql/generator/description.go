package generator

import (
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/jamesdphillips/graphql/language/ast"
)

// Fetch description from node if present; use 'self descriptive' if missing.
func getDescription(node ast.DescribableNode) string {
	desc := "self descriptive"
	if descNode := node.GetDescription(); descNode != nil {
		desc = descNode.Value
	}
	return desc
}

func genDescription(node ast.DescribableNode) jen.Code {
	desc := getDescription(node)
	jen.Lit(desc)
}

// To appease the linter ensure that the the description of the scalar begins
// with the name of the resulting method.
func genTypeComment(name, desc string) jen.Code {
	if hasPrefix := strings.HasPrefix(desc, name); !hasPrefix {
		desc = name + " " + desc
	}
	return desc
}
