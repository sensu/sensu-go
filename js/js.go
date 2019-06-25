package js

import (
	"fmt"
	"io"
	"sync"

	time "github.com/echlebek/timeproxy"
	"github.com/robertkrimen/otto"
	"github.com/robertkrimen/otto/parser"
	"github.com/sirupsen/logrus"
)

var logger = logrus.WithFields(logrus.Fields{
	"component": "filtering",
})

var ottoCache *vmCache
var ottoOnce sync.Once

type JavascriptAssets interface {
	Key() string
	Scripts() (map[string]io.ReadCloser, error)
}

// SyntaxError is returned when a javascript expression could not be parsed.
type SyntaxError string

func (s SyntaxError) Error() string {
	return string(s)
}

// NewSyntaxError creates a new SyntaxError.
func NewSyntaxError(err string, args ...interface{}) SyntaxError {
	return SyntaxError(fmt.Sprintf(err, args...))
}

// ParseExpressions parses each JS expression and returns the first error that
// is encountered, or nil.
func ParseExpressions(expressions []string) error {
	for i, expr := range expressions {
		_, err := parser.ParseFile(nil, "", expr, 0)
		if err != nil {
			return NewSyntaxError("syntax error in expression %d: %s", i, err)
		}
	}
	return nil
}

func newOttoVM(assets JavascriptAssets) (*otto.Otto, error) {
	ottoOnce.Do(func() {
		ottoCache = newVMCache()
	})
	key := ""
	if assets != nil {
		key = assets.Key()
	}
	vm := ottoCache.Acquire(key)
	if vm != nil {
		return vm, nil
	}
	vm = otto.New()
	if err := addTimeFuncs(vm); err != nil {
		return nil, err
	}
	if assets != nil {
		if err := addAssets(vm, assets); err != nil {
			return nil, err
		}
	}
	ottoCache.Init(key, vm)
	return ottoCache.Acquire(key), nil
}

func addAssets(vm *otto.Otto, assets JavascriptAssets) error {
	scripts, err := assets.Scripts()
	if err != nil {
		return err
	}
	defer func() {
		for _, script := range scripts {
			_ = script.Close()
		}
	}()
	for name, script := range scripts {
		if _, err := vm.Eval(script); err != nil {
			return fmt.Errorf("error evaluating %s: %s", name, err)
		}
	}
	return nil
}

func addTimeFuncs(vm *otto.Otto) error {
	funcs := map[string]interface{}{
		// hour returns the hour within the day
		"hour": func(args ...interface{}) interface{} {
			if len(args) == 0 {
				return 0
			}
			t := time.Unix(toInt64(args[0]), 0).UTC()
			return t.Hour()
		},
		// weekday returns the number representation of the day of the week, where
		// Sunday = 0
		"weekday": func(args ...interface{}) interface{} {
			if len(args) == 0 {
				return 0
			}
			t := time.Unix(toInt64(args[0]), 0).UTC()
			return t.Weekday()
		},
	}
	for k, v := range funcs {
		if err := vm.Set(k, v); err != nil {
			return err
		}
	}
	return nil
}

// Evaluate evaluates the javascript expression with parameters applied.
// If scripts is non-nil, then the scripts will be evaluated in the
// expression's runtime context before the expression is evaluated.
func Evaluate(expr string, parameters interface{}, assets JavascriptAssets) (bool, error) {
	jsvm, err := newOttoVM(assets)
	if err != nil {
		return false, err
	}
	if params, ok := parameters.(map[string]interface{}); ok {
		for name, value := range params {
			if err := jsvm.Set(name, value); err != nil {
				return false, err
			}
		}
	}
	value, err := jsvm.Run(expr)
	if err != nil {
		return false, err
	}
	return value.ToBoolean()
}

// EntityFilterResult is returned by EvaluateEntityFilters
type EntityFilterResult struct {
	Value bool
	Err   error
}

// MatchEntities compiles the expressions supplied, and applies each
// one of them to each entity supplied. On the first match, success is recorded
// and the evaluator moves on to the next entity. A slice of bools is returned
// that is the same length as the slice of entities supplied, indicating
// match success or failure.
//
// Errors are reported by logging only, with the log level determined by the
// severity of the error. Syntax and type errors are reported at error level,
// while attribute lookup errors are reported at debug level.
//
// If the function cannot set up a javascript VM, or has issues setting vars,
// then the function returns a nil slice and a non-nil error.
func MatchEntities(expressions []string, entities []interface{}) ([]bool, error) {
	jsvm, err := newOttoVM(nil)
	if err != nil {
		return nil, fmt.Errorf("error evaluating entity filters: %s", err)
	}
	scripts := make([]*otto.Script, 0, len(expressions))
	for _, expr := range expressions {
		script, err := jsvm.Compile("", expr)
		if err != nil {
			logger.WithError(err).Errorf("syntax error in script (%s)", expr)
			continue
		}
		scripts = append(scripts, script)
	}
	results := make([]bool, 0, len(entities))
	for _, entity := range entities {
		if err := jsvm.Set("entity", entity); err != nil {
			return nil, fmt.Errorf("error evaluating entity filters: %s", err)
		}
		var filtered bool
		for _, script := range scripts {
			result, err := jsvm.Run(script)
			if err != nil {
				logger.WithError(err).Debugf("error executing entity filter (%s)", script.String())
				filtered = false
				break
			}
			matches, err := result.ToBoolean()
			if err != nil {
				logger.WithError(err).Errorf("entity filter did not return bool (%s)", script.String())
				filtered = false
				break
			}
			if !matches {
				filtered = false
				break
			}
			// Mark the entity as filtered, but continue with the next script
			// (expression) until it went through all filters
			filtered = true
		}

		// At this point, the entity will be marked as filtered only if matched all
		// the expressions
		results = append(results, filtered)
	}

	return results, nil
}
