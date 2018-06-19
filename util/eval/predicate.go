package eval

import (
	"fmt"
	"time"

	"github.com/sensu/govaluate"
)

// SyntaxError is used to indicate that an expression could not be parsed.
type SyntaxError string

func (s SyntaxError) Error() string {
	return string(s)
}

// TypeError is used to indicate that a predicate expression did not evaluate
// to a bool.
type TypeError string

func (s TypeError) Error() string {
	return string(s)
}

// Predicate defines a group of logical expressions that can be used for
// in-memory filtering of resources.
type Predicate struct {
	expression *govaluate.EvaluableExpression
}

// NewPredicate initiailizes new predicate given expression.
func NewPredicate(expression string) (*Predicate, error) {
	funcs := expressionFunctions()

	expr, err := govaluate.NewEvaluableExpressionWithFunctions(expression, funcs)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the expression: %s", err.Error())
	}
	return &Predicate{expr}, nil
}

// Eval performs the evaluation of the given expression with provided
// parameters. An error is returned if it could not evaluate the expression with
// the provided parameters
func (p *Predicate) Eval(parameters govaluate.Parameters) (bool, error) {
	result, err := p.expression.Eval(parameters)
	if err != nil {
		return false, fmt.Errorf("failed to evaluate the expression: %s", err.Error())
	}

	match, ok := result.(bool)
	if !ok {
		return false, TypeError(fmt.Sprintf("govaluate expression: want bool, got %t", result))
	}

	return match, nil
}

// Evaluate is funcationally the same as Eval with the exception that given map
// is automatically wrapped into a `govaluate.Parameters` structure.
func (p *Predicate) Evaluate(parameters map[string]interface{}) (bool, error) {
	if parameters == nil {
		return p.Eval(nil)
	}
	return p.Eval(govaluate.MapParameters(parameters))
}

// EvaluatePredicate provides conveinent method of evaluating a given predicate
// when expression will only be evaluated one time.
func EvaluatePredicate(expression string, parameters map[string]interface{}) (bool, error) {
	predicate, err := NewPredicate(expression)
	if err != nil {
		return false, SyntaxError(err.Error())
	}
	return predicate.Evaluate(parameters)
}

// ValidateStatements ensure that the given statements can be parsed
// successfully and, optionally, that it does not contain any modifier tokens.
func ValidateStatements(statements []string, forbidModifier bool) error {
	funcs := expressionFunctions()

	for _, statement := range statements {
		exp, err := govaluate.NewEvaluableExpressionWithFunctions(statement, funcs)
		if err != nil {
			return SyntaxError(fmt.Sprintf("syntax error: %s (%s)", err, statement))
		}

		// We can optionally forbid modifier tokens if we believe an expression has
		// no reason to use a modifier operator in its context (e.g. assets
		// filters). By doing so, we can detect expressions that could possibly
		// evaluate to something else than a boolean value, and return an error
		// before saving that expression.
		if forbidModifier {
			// Ensure we don't have a modifier tokens (+, -, /, *, **, &, etc.)
			for _, token := range exp.Tokens() {
				if token.Kind == govaluate.MODIFIER {
					return fmt.Errorf("forbidden modifier tokens in statement '%s'", statement)
				}
			}
		}
	}

	return nil
}

func expressionFunctions() map[string]govaluate.ExpressionFunction {
	return map[string]govaluate.ExpressionFunction{
		// hour returns the hour within the day
		"hour": func(args ...interface{}) (interface{}, error) {
			t := time.Unix(int64(args[0].(float64)), 0).UTC()
			return float64(t.Hour()), nil
		},
		// weekday returns the number representation of the day of the week, where
		// Sunday = 0
		"weekday": func(args ...interface{}) (interface{}, error) {
			t := time.Unix(int64(args[0].(float64)), 0).UTC()
			return float64(t.Weekday()), nil
		},
	}
}
