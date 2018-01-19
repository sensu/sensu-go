package generator

import (
	"strings"

	"github.com/dave/jennifer/jen"
	"github.com/graphql-go/graphql/language/ast"
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
	return jen.Lit(desc)
}

// To appease the linter ensure that the the description of the type begins
// with the name of the resulting method.
func genTypeComment(name, desc string) string {
	if hasPrefix := strings.HasPrefix(desc, name); !hasPrefix {
		desc = name + " " + desc
	}
	return desc
}

// To appease the linter ensure that the the description of the field begins
// with the name of the resulting method. Includes conventional deprecation
// notice if applicable.
func genFieldComment(name, desc, depr string) string {
	fName := strings.Title(name)
	if hasPrefix := strings.HasPrefix(desc, fName); !hasPrefix {
		desc = fName + " - " + desc
	}
	if depr != "" {
		desc = desc + "\n\nDeprecated: " + depr
	}
	return desc
}
