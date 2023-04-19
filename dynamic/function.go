package dynamic

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/robertkrimen/otto"
)

// Function wraps a function for execution in Javascript. It returns an
// object suitable for calling from a Javascript interpreter.
//
// Function can take either a Go function, or a string that evaluates to a
// Javascript function.
//
// If fn returns an error as its last return value, then Function will convert
// that error into a Javscript exception, if the error is non-nil.
//
// If fn takes a context as its first argument, then Function will remove that
// argument from the returned Javascript function and implicitly inject the
// provided context.
//
// Function will panic if called with an fn that does not have a reflect type
// of 'Func'.
//
// A function like GetEvents(context.Context) ([]*corev2.Event, error) will get
// translated into something like GetEvents() []*corev2.Event.
//
// The result will be a raw Go object.
func Function(ctx context.Context, vm *otto.Otto, fn interface{}) interface{} {
	value := reflect.ValueOf(fn)
	typ := reflect.TypeOf(fn)
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	ctxType := reflect.TypeOf((*context.Context)(nil)).Elem()
	ctxValue := reflect.ValueOf(ctx)

	switch typ.Kind() {
	case reflect.Func:
		return func(args ...interface{}) (result interface{}) {
			defer func() {
				if e := recover(); e != nil {
					s := fmt.Sprintf("%s", e)
					result = otto.UndefinedValue()
					if strings.HasPrefix(s, "reflect: ") {
						s = strings.TrimPrefix(s, "reflect: ")
						panic(vm.MakeTypeError(s))
					} else {
						panic(vm.MakeCustomError("Error", s))
					}
				}
			}()
			numArgs := typ.NumIn()
			argValues := make([]reflect.Value, numArgs)
			argsIdx := 0
			for i := 0; i < numArgs; i++ {
				if i == 0 && typ.In(0).Implements(ctxType) {
					argValues[0] = ctxValue
					continue
				}
				if argsIdx >= len(args) || args[argsIdx] == nil {
					argValues[i] = reflect.New(typ.In(i)).Elem()
				} else {
					argValues[i] = reflect.ValueOf(args[argsIdx])
				}
				argsIdx++
			}
			callResults := value.Call(argValues)
			if len(callResults) == 0 {
				return otto.UndefinedValue()
			}
			if !typ.Out(typ.NumOut() - 1).Implements(errorType) {
				return toInterface(callResults)
			}
			errVal := callResults[len(callResults)-1].Interface()
			if errVal != nil {
				err := errVal.(error)
				panic(vm.MakeCustomError("sensu", err.Error()))
			}
			return toInterface(callResults[:len(callResults)-1])
		}
	case reflect.String:
		funcVal, err := vm.Eval(fn)
		if err != nil {
			return func(args ...interface{}) (result interface{}) {
				panic(vm.MakeTypeError(err.Error()))
			}
		}
		return funcVal
	}
	panic(fmt.Sprintf("call with non-func type %s", typ.Kind()))
}

func toInterface(values []reflect.Value) interface{} {
	switch len(values) {
	case 0:
		return otto.UndefinedValue()
	case 1:
		return values[0].Interface()
	default:
		result := make([]interface{}, len(values))
		for i := range result {
			result[i] = values[i].Interface()
		}
		return result
	}
}
