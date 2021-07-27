package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/js"
	"github.com/sensu/sensu-go/types/dynamic"
)

const pipelineRoleName = "system:pipeline"

// PipelineFilterFuncs gets patched by enterprise sensu-go
var PipelineFilterFuncs map[string]interface{}

type FilterExecutionEnvironment struct {
	// Funcs are a list of named functions to be supplied to the JS environment.
	Funcs map[string]interface{}

	// Assets are a set of javascript assets to be evaluated before filter
	// execution.
	Assets js.JavascriptAssets

	// Event is the Sensu event to be supplied to the filter execution environment.
	Event interface{}
}

func (f *FilterExecutionEnvironment) Eval(ctx context.Context, expression string) (bool, error) {
	// these claims set up authorization credentials for the event client.
	// the actual roles and role bindings are generated when namespaces are
	// created, for the purposes of the system.
	claims := &corev2.Claims{
		Groups: []string{pipelineRoleName},
		Provider: corev2.AuthProviderClaims{
			ProviderID:   "system",
			ProviderType: "basic",
			UserID:       pipelineRoleName,
		},
	}
	ctx = context.WithValue(ctx, corev2.ClaimsKey, claims)
	parameters := map[string]interface{}{}
	var assets js.JavascriptAssets
	if f != nil {
		parameters["event"] = f.Event
		assets = f.Assets
	}
	var result bool
	err := js.WithOttoVM(assets, func(vm *otto.Otto) (err error) {
		if f != nil {
			funcs := make(map[string]interface{}, len(f.Funcs))
			for k, v := range f.Funcs {
				funcs[k] = dynamic.Function(ctx, vm, v)
			}
			parameters["sensu"] = funcs
		}
		result, err = js.EvalPredicateWithVM(vm, parameters, expression)
		return err
	})
	return result, err
}

type MutatorExecutionEnvironment struct {
	// Env is the "environment" of the mutator
	Env []string

	// Event is the event to be mutated. Unlike with filters, the event is not
	// serialized.
	Event *corev2.Event

	// Assets are a set of javascript assets to be evaluated before mutator
	// execution.
	Assets js.JavascriptAssets

	// Timeout specifies how long to allow the mutator to run for
	Timeout time.Duration
}

func parseEnv(vars []string) map[string]string {
	result := make(map[string]string, len(vars))
	for _, kv := range vars {
		idx := strings.Index(kv, "=")
		if idx <= 0 {
			continue
		}
		result[kv[:idx]] = kv[idx+1:]
	}
	return result
}

var halt = errors.New("halt")

func (m *MutatorExecutionEnvironment) Eval(ctx context.Context, expression string) ([]byte, error) {
	env := parseEnv(m.Env)

	var assets js.JavascriptAssets

	parameters := make(map[string]interface{})
	parameters["event"] = m.Event
	assets = m.Assets
	var result []byte

	err := js.WithOttoVM(assets, func(vm *otto.Otto) (err error) {
		for k, v := range parameters {
			if err := vm.Set(k, v); err != nil {
				return err
			}
		}
		for k, v := range env {
			if err := vm.Set(k, v); err != nil {
				return err
			}
		}
		vm.Interrupt = make(chan func(), 1)
		defer func() {
			if e := recover(); e != nil && e == halt {
				err = errors.New("mutator timeout reached, execution halted")
			} else if e != nil {
				panic(e)
			}
		}()
		done := make(chan struct{})
		if m.Timeout > 0 {
			go func() {
				select {
				case <-time.After(m.Timeout):
					vm.Interrupt <- func() {
						panic(halt)
					}
				case <-done:
				}
			}()
		}
		value, err := vm.Run(fmt.Sprintf("(function () { %s }())", expression))
		if err != nil {
			return err
		}
		close(done)
		if value.IsUndefined() || value.IsNull() {
			result, err = json.Marshal(m.Event)
		} else if value.IsString() {
			str, _ := value.ToString()
			result = []byte(str)
		} else {
			err = fmt.Errorf("bad mutator result: got %q, want string or undefined", value.Class())
		}
		return err
	})
	return result, err
}
