package mutator

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/robertkrimen/otto"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/js"
	"github.com/sensu/sensu-go/util/environment"
	utillogging "github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

var (
	halt = errors.New("halt")

	builtInMutatorNames = []string{
		"json",
		"only_check_output",
	}
)

type Legacy struct {
	AssetGetter            asset.Getter
	Executor               command.Executor
	SecretsProviderManager *secrets.ProviderManager
	Store                  store.Store
	StoreTimeout           time.Duration
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

// CanMutate determines whether the Legacy mutator can mutate the resource being
// referenced.
func (l *Legacy) CanMutate(ctx context.Context, ref *corev2.ResourceReference) bool {
	if ref.APIVersion == "core/v2" && ref.Type == "Mutator" {
		for _, name := range builtInMutatorNames {
			if ref.Name == name {
				return false
			}
		}
		return true
	}
	return false
}

// Mutate mutates (transforms) a Sensu event into a serialized format (byte
// slice) to be provided to a Sensu event handler.
func (l *Legacy) Mutate(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event) ([]byte, error) {
	// Prepare log entry
	// TODO: add pipeline & pipeline workflow names to fields
	fields := utillogging.EventFields(event, false)

	ctx = context.WithValue(ctx, corev2.NamespaceKey, event.Entity.Namespace)
	tctx, cancel := context.WithTimeout(ctx, l.StoreTimeout)
	defer cancel()

	mutator, err := l.Store.GetMutatorByName(tctx, ref.Name)
	if err != nil {
		// Warning: do not wrap this error
		logger.WithFields(fields).WithError(err).Error("failed to retrieve mutator")
		return nil, err
	}

	var eventData []byte

	var assets asset.RuntimeAssetSet
	if len(mutator.RuntimeAssets) > 0 {
		logger.WithFields(fields).Debug("fetching assets for mutator")

		// Fetch and install all assets required for handler execution
		matchedAssets := asset.GetAssets(ctx, l.Store, mutator.RuntimeAssets)

		var err error
		assets, err = asset.GetAll(ctx, l.AssetGetter, matchedAssets)
		if err != nil {
			logger.WithFields(fields).WithError(err).Error("failed to retrieve assets for mutator")
		}
	}

	if mutator.Type == "" || mutator.Type == corev2.PipeMutator {
		eventData, err = l.pipeMutator(ctx, mutator, event, assets)
	} else if mutator.Type == corev2.JavascriptMutator {
		eventData, err = l.javascriptMutator(ctx, mutator, event, assets)
	}

	if err != nil {
		logger.WithFields(fields).WithError(err).Error("failed to mutate the event")
		return nil, err
	}

	return eventData, nil
}

// pipeMutator fork/executes a child process for a Sensu mutator command, writes
// the JSON encoding of the Sensu event to it via STDIN, and captures the
// command output (STDOUT/ERR) to be used as the mutated event data for a Sensu
// event handler.
func (l *Legacy) pipeMutator(ctx context.Context, mutator *corev2.Mutator, event *corev2.Event, assets asset.RuntimeAssetSet) ([]byte, error) {
	ctx = corev2.SetContextFromResource(ctx, mutator)

	// Prepare log entry
	fields := logrus.Fields{
		"namespace": mutator.Namespace,
		"mutator":   mutator.Name,
		"assets":    mutator.RuntimeAssets,
	}

	secrets, err := l.SecretsProviderManager.SubSecrets(ctx, mutator.Secrets)
	if err != nil {
		logger.WithFields(fields).WithError(err).Error("failed to retrieve secrets for mutator")
		return nil, err
	}

	// Prepare environment variables
	env := environment.MergeEnvironments(os.Environ(), mutator.EnvVars, secrets)

	mutatorExec := command.ExecutionRequest{}
	mutatorExec.Command = mutator.Command
	mutatorExec.Timeout = int(mutator.Timeout)
	mutatorExec.Env = env

	eventData, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	mutatorExec.Input = string(eventData[:])

	// Only add assets to execution context if handler requires them
	if assets != nil {
		mutatorExec.Env = environment.MergeEnvironments(os.Environ(), assets.Env(), mutator.EnvVars, secrets)
	}

	result, err := l.Executor.Execute(context.Background(), mutatorExec)

	fields["status"] = result.Status
	fields["output"] = result.Output
	if err != nil {
		logger.WithFields(fields).WithError(err).Error("failed to execute event pipe mutator")
		return nil, err
	} else if result.Status != 0 {
		logger.WithFields(fields).Error("failure in event pipe mutator execution")
		return nil, errors.New("pipe mutator execution returned non-zero exit status")
	}

	logger.WithFields(fields).Debug("pipe event mutator executed")

	return []byte(result.Output), nil
}

func (l *Legacy) javascriptMutator(ctx context.Context, mutator *corev2.Mutator, event *corev2.Event, assets js.JavascriptAssets) ([]byte, error) {
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
