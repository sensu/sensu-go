package mutator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/js"
	"github.com/sirupsen/logrus"
)

var (
	halt = errors.New("halt")
)

// JavascriptAdapter is a mutator adapter which mutates an event using
// javascript.
type JavascriptAdapter struct {
	AssetGetter  asset.Getter
	Store        store.Store
	StoreTimeout time.Duration
}

// Name returns the name of the mutator adapter.
func (j *JavascriptAdapter) Name() string {
	return "JavascriptAdapter"
}

// CanMutate determines whether JavascriptAdapter can mutate the resource
// being referenced.
// TODO: update this function if/when we implement a proper JavascriptMutator type
func (j *JavascriptAdapter) CanMutate(ref *corev2.ResourceReference) bool {
	return false
}

// Mutate will mutate the event and return it as bytes.
// TODO: update this function if/when we implement a proper JavascriptMutator type
func (j *JavascriptAdapter) Mutate(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event) ([]byte, error) {
	// eventData, err := j.run(ctx, mutator, event, assets)
	// if err != nil {
	// 	return nil, err
	// }
	return nil, fmt.Errorf("mutator adapter cannot be used directly at this time")
}

func (j *JavascriptAdapter) run(ctx context.Context, mutator *corev2.Mutator, event *corev2.Event, assets js.JavascriptAssets) ([]byte, error) {
	ctx = corev2.SetContextFromResource(context.Background(), mutator)

	// Use a context with a timeout if the mutator has a timeout set
	var cancel context.CancelFunc
	if mutator.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, time.Second*time.Duration(mutator.Timeout))
		defer cancel()
	}
	// Prepare log entry
	fields := logrus.Fields{
		"namespace": mutator.Namespace,
		"mutator":   mutator.Name,
		"assets":    mutator.RuntimeAssets,
	}

	// Guard against nil metadata labels and annotations to improve the user
	// experience of querying these them.
	if event.ObjectMeta.Annotations == nil {
		event.ObjectMeta.Annotations = make(map[string]string)
	}
	if event.ObjectMeta.Labels == nil {
		event.ObjectMeta.Labels = make(map[string]string)
	}
	if event.Check.ObjectMeta.Annotations == nil {
		event.Check.ObjectMeta.Annotations = make(map[string]string)
	}
	if event.Check.ObjectMeta.Labels == nil {
		event.Check.ObjectMeta.Labels = make(map[string]string)
	}
	if event.Entity.ObjectMeta.Annotations == nil {
		event.Entity.ObjectMeta.Annotations = make(map[string]string)
	}
	if event.Entity.ObjectMeta.Labels == nil {
		event.Entity.ObjectMeta.Labels = make(map[string]string)
	}

	env := MutatorExecutionEnvironment{
		Event:   event,
		Env:     mutator.EnvVars,
		Timeout: time.Duration(mutator.Timeout) * time.Second,
		Assets:  assets,
	}

	logger.WithFields(fields).Debug("javascript event mutator executed")

	return env.Eval(ctx, mutator.Eval)
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
