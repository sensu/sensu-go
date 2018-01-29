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

// ValidateStatements ensure that the given statements can be parsed
// successfully and that it does not contain any modifier tokens.
func ValidateStatements(statements []string) error {
	for _, statement := range statements {
		exp, err := govaluate.NewEvaluableExpression(statement)
		if err != nil {
			return fmt.Errorf("invalid statement '%s': %s", statement, err.Error())
		}

		// Do not allow modifier tokens (eg. +, -, /, *, **, &, etc.)
		for _, token := range exp.Tokens() {
			if token.Kind == govaluate.MODIFIER {
				return fmt.Errorf("forbidden modifier tokens in statement '%s'", statement)
			}
		}
	}

	return nil
}
