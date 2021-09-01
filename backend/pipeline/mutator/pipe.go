package mutator

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/util/environment"
	"github.com/sirupsen/logrus"
)

// PipeAdapter is a mutator adapter which fork/executes a child process for a
// Sensu mutator command, writes the JSON encoding of the Sensu event to it via
// STDIN, and captures the command output (STDOUT/ERR) to be used as the mutated
// event data for a Sensu event handler.
type PipeAdapter struct {
	AssetGetter            asset.Getter
	Executor               command.Executor
	SecretsProviderManager *secrets.ProviderManager
	Store                  store.Store
	StoreTimeout           time.Duration
}

// Name returns the name of the mutator adapter.
func (p *PipeAdapter) Name() string {
	return "PipeAdapter"
}

// CanMutate determines whether PipeAdapter can mutate the resource being
// referenced.
// TODO: update this function if/when we implement a proper PipeMutator type
func (p *PipeAdapter) CanMutate(ctx context.Context, ref *corev2.ResourceReference) bool {
	return false
}

// Mutate will mutate the event and return it as bytes.
// TODO: update this function if/when we implement a proper PipeMutator type
func (p *PipeAdapter) Mutate(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event) ([]byte, error) {
	// eventData, err := p.run(ctx, mutator, event, assets)
	// if err != nil {
	// 	return nil, err
	// }
	return nil, errors.New("mutator adapter cannot be used directly at this time")
}

func (p *PipeAdapter) run(ctx context.Context, mutator *corev2.Mutator, event *corev2.Event, assets asset.RuntimeAssetSet) ([]byte, error) {
	ctx = corev2.SetContextFromResource(ctx, mutator)

	// Prepare log entry
	fields := logrus.Fields{
		"namespace": mutator.Namespace,
		"mutator":   mutator.Name,
		"assets":    mutator.RuntimeAssets,
	}

	secrets, err := p.SecretsProviderManager.SubSecrets(ctx, mutator.Secrets)
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

	result, err := p.Executor.Execute(ctx, mutatorExec)

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
