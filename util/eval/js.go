package eval

import (
	"fmt"

	"github.com/robertkrimen/otto"
	"github.com/robertkrimen/otto/parser"
)

var jsvm = otto.New()

// ValidateJSStatements parses each JS statement and returns the first error
// that is encountered, or nil.
func ValidateJSStatements(stmt []string) error {
	for i, statement := range stmt {
		_, err := parser.ParseFile(nil, "", statement, 0)
		if err != nil {
			return fmt.Errorf("error parsing statement %d: %s", i, err)
		}
	}
	return nil
}

// EvaluateJSExpression evaluates the javascript expression with parameters applied
func EvaluateJSExpression(expr string, parameters map[string]interface{}) (bool, error) {
	for name, value := range parameters {
		if err := jsvm.Set(name, value); err != nil {
			return false, err
		}
	}
	value, err := jsvm.Run(expr)
	if err != nil {
		return false, err
	}
	return value.ToBoolean()
}
