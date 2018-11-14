package js

import (
	"fmt"
	"io"
	"sync"

	time "github.com/echlebek/timeproxy"
	"github.com/robertkrimen/otto"
	"github.com/robertkrimen/otto/parser"
)

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

func releaseOttoVM(vm *otto.Otto, assets JavascriptAssets) {
	key := ""
	if assets != nil {
		key = assets.Key()
	}
	ottoCache.Release(key, vm)
}

// Evaluate evaluates the javascript expression with parameters applied.
// If scripts is non-nil, then the scripts will be evaluated in the
// expression's runtime context before the expression is evaluated.
func Evaluate(expr string, parameters interface{}, assets JavascriptAssets) (bool, error) {
	jsvm, err := newOttoVM(assets)
	if err != nil {
		return false, err
	}
	defer releaseOttoVM(jsvm, assets)
	if params, ok := parameters.(map[string]interface{}); ok {
		defer func() {
			for name := range params {
				_ = jsvm.Set(name, otto.UndefinedValue())
			}
		}()
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
