package graphql

import (
	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/gqlerrors"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/kinds"
	"github.com/graphql-go/graphql/language/visitor"
)

const (
	MaxQueryNodeDepth = 5
)

func reportError(context *graphql.ValidationContext, message string, nodes []ast.Node) (string, interface{}) {
	context.ReportError(gqlerrors.NewError(message, nodes, "", nil, []int{}, nil))
	return visitor.ActionNoChange, nil
}

// validate max depth of a GraphQL query with a depthLimit
func validateMaxDepth(context *graphql.ValidationContext, node ast.Node, currentDepth int, depthLimit int) int {
	// end recursion early if error reported
	if errors := context.Errors(); len(errors) > 0 {
		return -1
	}

	if currentDepth > depthLimit {
		reportError(context, "Max depth exceeded", []ast.Node{node})
		return -1
	}

	var selectionSet *ast.SelectionSet

	switch node.GetKind() {
	case kinds.Field:
		selectionSet = node.(*ast.Field).GetSelectionSet()
		if selectionSet != nil {
			selections := selectionSet.Selections
			maxDepth := currentDepth
			for _, selection := range selections {
				nextDepth := validateMaxDepth(context, selection.(ast.Node), currentDepth+1, depthLimit)
				if nextDepth > maxDepth {
					maxDepth = nextDepth
				}
			}
			return maxDepth
		}
		return 0
	case kinds.FragmentSpread:
		fragName := node.(*ast.FragmentSpread).Name.Value
		spreadFragment := context.Fragment(fragName)
		if spreadFragment != nil {
			return validateMaxDepth(context, spreadFragment, currentDepth, depthLimit)
		}
		return 0
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
			nextDepth := validateMaxDepth(context, selection.(ast.Node), currentDepth, depthLimit)
			if nextDepth > maxDepth {
				maxDepth = nextDepth
			}
		}
		return maxDepth
	}
	return 0
}

func (rule *maxDepthRule) maxDepthVisitorOptions() *visitor.VisitorOptions {
	return &visitor.VisitorOptions{
		KindFuncMap: map[string]visitor.NamedVisitFuncs{
			kinds.OperationDefinition: {
				Kind: func(p visitor.VisitFuncParams) (string, interface{}) {
					node := p.Node.(ast.Node)
					if node != nil {
						validateMaxDepth(rule.context, node, 0, rule.depthLimit)
					}
					return visitor.ActionNoChange, nil
				},
			},
		},
	}
}

type maxDepthRule struct {
	context    *graphql.ValidationContext
	depthLimit int
}

func newMaxDepthRule(context *graphql.ValidationContext, depthLimit int) *maxDepthRule {
	return &maxDepthRule{
		context:    context,
		depthLimit: depthLimit,
	}
}

func MaxDepthRule(depthLimit int) graphql.ValidationRuleFn {
	rule := &maxDepthRule{
		depthLimit: depthLimit,
	}
	return rule.validateRule
}

func (r *maxDepthRule) validateRule(context *graphql.ValidationContext) *graphql.ValidationRuleInstance {
	rule := newMaxDepthRule(context, r.depthLimit)
	return &graphql.ValidationRuleInstance{VisitorOpts: rule.maxDepthVisitorOptions()}
}

type UnauthedValidatorOpts struct {
	DepthLimit int
}

func ProvideUnauthedValidators(opts UnauthedValidatorOpts) []graphql.ValidationRuleFn {
	rules := []graphql.ValidationRuleFn{MaxDepthRule(opts.DepthLimit)}
	return rules
}
