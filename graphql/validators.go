package graphql

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/kinds"
	"github.com/graphql-go/graphql/language/visitor"
)

const (
	MaxQueryDepthLimit = 15
)

type maxDepthRule struct {
	context    *graphql.ValidationContext
	depth      int
	depthLimit int
}

func MandatoryValidators() []graphql.ValidationRuleFn {
	return []graphql.ValidationRuleFn{MaxDepthRule(MaxQueryDepthLimit)}
}

func MaxDepthRule(depthLimit int) graphql.ValidationRuleFn {
	rule := newMaxDepthRule(depthLimit)
	return rule.maxDepthRuleWithContext
}

// provide MaxDepthRule with depth limit
func newMaxDepthRule(depthLimit int) *maxDepthRule {
	rule := &maxDepthRule{
		depthLimit: depthLimit,
	}
	return rule
}

func (r *maxDepthRule) maxDepthRuleWithContext(context *graphql.ValidationContext) *graphql.ValidationRuleInstance {
	rule := &maxDepthRule{
		context:    context,
		depthLimit: r.depthLimit,
	}
	return &graphql.ValidationRuleInstance{VisitorOpts: rule.maxDepthVisitorOptions()}
}

func (rule *maxDepthRule) maxDepthVisitorOptions() *visitor.VisitorOptions {
	return &visitor.VisitorOptions{
		KindFuncMap: map[string]visitor.NamedVisitFuncs{
			kinds.OperationDefinition: {
				Kind: func(p visitor.VisitFuncParams) (string, interface{}) {
					node := p.Node.(ast.Node)
					if node != nil {
						maxDepth := validateMaxDepth(rule.context, node, 1, rule.depthLimit, nil)
						rule.depth = maxDepth
					}
					return visitor.ActionNoChange, nil
				},
			},
		},
	}
}

// validate max depth of a GraphQL query with a depthLimit
func validateMaxDepth(
	context *graphql.ValidationContext,
	node ast.Node,
	currentDepth int,
	depthLimit int,
	visited map[*ast.FragmentDefinition]bool,
) int {
	// end recursion early if error reported
	if errors := context.Errors(); len(errors) > 0 {
		return -1
	}

	if currentDepth > depthLimit {
		reportError(context, "Max depth exceeded", []ast.Node{node})
		return -1
	}

	// keep map of visited fragment spreads to prevent infinite loop
	if visited == nil {
		visited = map[*ast.FragmentDefinition]bool{}
	}

	var selectionSet *ast.SelectionSet

	switch node.GetKind() {
	case kinds.Field:
		selectionSet = node.(*ast.Field).GetSelectionSet()
		if selectionSet != nil {
			selections := selectionSet.Selections
			maxDepth := currentDepth
			for _, selection := range selections {
				nextDepth := validateMaxDepth(context, selection.(ast.Node), currentDepth+1, depthLimit, visited)
				if nextDepth > maxDepth {
					maxDepth = nextDepth
				}
			}
			return maxDepth
		}
		return currentDepth
	case kinds.FragmentSpread:
		fragName := node.(*ast.FragmentSpread).Name.Value
		fragment := context.Fragment(fragName)
		if fragment == nil || visited[fragment] {
			return currentDepth
		}
		visited[fragment] = true
		// fragment spreads don't increase the depth
		return validateMaxDepth(context, fragment, currentDepth, depthLimit, visited)
	case kinds.InlineFragment:
		selectionSet = node.(*ast.InlineFragment).GetSelectionSet()
	case kinds.FragmentDefinition:
		selectionSet = node.(*ast.FragmentDefinition).GetSelectionSet()
	case kinds.OperationDefinition:
		selectionSet = node.(*ast.OperationDefinition).GetSelectionSet()
	}

	if selectionSet != nil {
		selections := selectionSet.Selections
		maxDepth := currentDepth
		for _, selection := range selections {
			// inline fragments, fragment definitions and operation definitions don't increase the depth
			nextDepth := validateMaxDepth(context, selection.(ast.Node), currentDepth, depthLimit, visited)
			if nextDepth > maxDepth {
				maxDepth = nextDepth
			}
		}
		return maxDepth
	}
	return currentDepth
}

func reportError(context *graphql.ValidationContext, message string, nodes []ast.Node) {
	context.ReportError(gqlerrors.NewError(message, nodes, "", nil, []int{}, nil))
}
