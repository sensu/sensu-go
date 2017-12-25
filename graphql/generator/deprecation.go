package generator

import "github.com/jamesdphillips/graphql/language/ast"

// Fetch deprecation reason given set of directives; default to empty string.
// Specification: https://github.com/facebook/graphql/blob/398e443983724463b8474b12a260fba31c19c2a9/spec/Section%203%20--%20Type%20System.md#deprecated
func fetchDeprecationReason(ds []*ast.Directive) string {
	// Determine if deprecated directive was given.
	var directive *ast.Directive
	for _, d := range ds {
		if d.Name.Value != "deprecated" {
			continue
		}
		directive = d
	}

	// If deprecated directive is not present return an empty string.
	if directive == nil {
		return ""
	}

	// Iterate through arguments for reason and return if given.
	for _, arg := range directive.Arguments {
		if arg.Name.Value != "reason" {
			continue
		}
		val := arg.Value.GetValue()
		if reason, ok := val.(string); ok {
			return reason
		}
		logger.WithField("value", val).Error("given deprecation reason of unexpected type")
		break
	}

	// As per GraphQL spec fallback to 'No longer supported'.
	return "No longer supported"
}
