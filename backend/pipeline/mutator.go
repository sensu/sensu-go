package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/js"
	"github.com/sensu/sensu-go/util/environment"
	"github.com/sirupsen/logrus"
)

type Mutator interface {
	CanMutate(context.Context, *corev2.ResourceReference) bool
	Mutate(context.Context, *corev2.ResourceReference, *corev2.Event) (*corev2.Event, error)
}

func (p *Pipeline) getMutatorForResource(ctx context.Context, ref *corev2.ResourceReference) (Mutator, error) {
	for _, mutator := range p.mutators {
		if mutator.CanMutate(ctx, ref) {
			return mutator, nil
		}
	}
	return nil, fmt.Errorf("no mutator processors were found that can mutate the resource: %s.%s = %s", ref.APIVersion, ref.Type, ref.Name)
}

func (p *Pipeline) processMutator(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event) (*corev2.Event, error) {
	mutator, err := p.getMutatorForResource(ctx, ref)
	if err != nil {
		return nil, err
	}

	return mutator.Mutate(ctx, ref, event)
}

func (p *Pipeline) mutateEvent(handler *corev2.Handler, event *corev2.Event) ([]byte, error) {

}

// jsonMutator produces the JSON encoding of the Sensu event. This
// mutator is used when a Sensu handler does not specify one.
func (p *Pipeline) jsonMutator(event *corev2.Event) ([]byte, error) {
	eventData, err := json.Marshal(event)

	if err != nil {
		return nil, err
	}

	return eventData, nil
}

// onlyCheckOutputMutator returns only the check output from the Sensu
// event. This mutator is considered to be "built-in" (1.x parity), it
// is most commonly used by tcp/udp handlers (e.g. influxdb). This
// mutator can probably be removed/replaced when 2.0 has extension
// support.
func (p *Pipeline) onlyCheckOutputMutator(event *corev2.Event) []byte {
	if event.HasCheck() {
		return []byte(event.Check.Output)
	}
	return nil
}

// pipeMutator fork/executes a child process for a Sensu mutator
// command, writes the JSON encoding of the Sensu event to it via
// STDIN, and captures the command output (STDOUT/ERR) to be used as
// the mutated event data for a Sensu event handler.
func (p *Pipeline) pipeMutator(mutator *corev2.Mutator, event *corev2.Event, assets asset.RuntimeAssetSet) ([]byte, error) {
	ctx := corev2.SetContextFromResource(context.Background(), mutator)
	// Prepare log entry
	fields := logrus.Fields{
		"namespace": mutator.Namespace,
		"mutator":   mutator.Name,
		"assets":    mutator.RuntimeAssets,
	}

	secrets, err := p.secretsProviderManager.SubSecrets(ctx, mutator.Secrets)
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

	result, err := p.executor.Execute(context.Background(), mutatorExec)

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

func (p *Pipeline) javascriptMutator(mutator *corev2.Mutator, event *corev2.Event, assets js.JavascriptAssets) ([]byte, error) {
	ctx := corev2.SetContextFromResource(context.Background(), mutator)
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
