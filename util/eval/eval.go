package eval

import (
	"fmt"

	"github.com/sensu/govaluate"
)

// Evaluate performs the evaluation of the given expression with provided
// parameters. An error is returned if it could not evaluate the expression with
// the provided parameters
func Evaluate(expression string, parameters map[string]interface{}) (bool, error) {
	expr, err := govaluate.NewEvaluableExpression(expression)
	if err != nil {
		return false, fmt.Errorf("failed to parse the expression: %s", err.Error())
	}

	result, err := expr.Evaluate(parameters)
	if err != nil {
		return false, fmt.Errorf("failed to evaluate the expression: %s", err.Error())
	}

	match, ok := result.(bool)
	if !ok {
		return false, fmt.Errorf("expression result was non-boolean value")
	}

	return match, nil
}
