package generator

import "github.com/jamesdphillips/graphql/language/ast"

// Fetch description from node if present; use 'self descriptive' if missing.
func fetchDescription(node ast.DescribableNode) string {
	desc := "self descriptive"
	if descNode := node.GetDescription(); descNode != nil {
		desc = descNode.Value
	}
	return desc
}
