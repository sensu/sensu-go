package sdk

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/robertkrimen/otto"
)

func TestCall(t *testing.T) {
	tests := []struct {
		Name     string
		Func     interface{}
		ExpError bool
		Exp      interface{}
		Args     []interface{}
	}{
		{
			Name: "no args, no return",
			Func: func() {},
			Exp:  otto.UndefinedValue(),
		},
		{
			Name: "no args, one return value",
			Func: func() interface{} { return "foo" },
			Exp:  "foo",
		},
		{
			Name: "no args, return value with nil error",
			Func: func() error { return nil },
			Exp:  otto.UndefinedValue(),
		},
		{
			Name:     "no args, return value with non-nil error",
			Func:     func() error { return errors.New("error") },
			ExpError: true,
		},
		{
			Name: "no args, multiple return values",
			Func: func() (interface{}, interface{}) { return "foo", "bar" },
			Exp:  []interface{}{"foo", "bar"},
		},
		{
			Name: "no args, two return values with nil error",
			Func: func() (interface{}, error) { return "foo", nil },
			Exp:  "foo",
		},
		{
			Name:     "no args, multiple return values, non-nil error",
			Func:     func() (interface{}, error) { return "", errors.New("error") },
			ExpError: true,
		},
		{
			Name: "one arg, no return",
			Func: func(interface{}) {},
			Exp:  otto.UndefinedValue(),
			Args: []interface{}{1},
		},
		{
			Name: "one arg, one return value",
			Func: func(a interface{}) interface{} { return a },
			Exp:  "foo",
			Args: []interface{}{"foo"},
		},
		{
			Name: "one arg, return value with nil error",
			Func: func(a interface{}) error { return nil },
			Exp:  otto.UndefinedValue(),
			Args: []interface{}{"foo"},
		},
		{
			Name:     "one arg, return value with non-nil error",
			Func:     func(a interface{}) error { return errors.New("error") },
			ExpError: true,
			Args:     []interface{}{"foo"},
		},
		{
			Name: "one arg, multiple return values",
			Func: func(a interface{}) (interface{}, interface{}) { return a, a },
			Exp:  []interface{}{"foo", "foo"},
			Args: []interface{}{"foo"},
		},
		{
			Name: "one arg, two return values with nil error",
			Func: func(a interface{}) (interface{}, error) { return a, nil },
			Exp:  "foo",
			Args: []interface{}{"foo"},
		},
		{
			Name:     "one arg, multiple return values, non-nil error",
			Func:     func(a interface{}) (interface{}, error) { return a, errors.New("error") },
			ExpError: true,
			Args:     []interface{}{"foo"},
		},
		{
			Name: "less than the supported number of args",
			Func: func(a interface{}) {},
			Args: nil,
			Exp:  otto.UndefinedValue(),
		},
		{
			Name: "more than the supported number of args",
			Func: func(a interface{}) interface{} { return a },
			Args: []interface{}{1, 2, 3},
			Exp:  1,
		},
		{
			Name:     "wrong type of args",
			Func:     func(a int) {},
			ExpError: true,
			Args:     []interface{}{"foo"},
		},
		{
			Name:     "panic in func",
			Func:     func() { panic("!") },
			ExpError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			defer func() {
				if e := recover(); e != nil {
					if !test.ExpError {
						t.Fatal(e)
					}
					return
				}
				if test.ExpError {
					t.Fatal("expected error")
				}
			}()
			vm := otto.New()
			callable := call(context.Background(), vm, test.Func)
			result := callable(test.Args...)
			if !test.ExpError {
				if got, want := result, test.Exp; !reflect.DeepEqual(got, want) {
					t.Fatalf("bad result: got %v, want %v", got, want)
				}
			}
		})
	}
}
