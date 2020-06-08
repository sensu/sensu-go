package pipeline

import (
	"context"
	"encoding/json"
	"errors"
	"os"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/util/environment"
	utillogging "github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

// mutateEvent mutates (transforms) a Sensu event into a serialized
// format (byte slice) to be provided to a Sensu event handler.
func (p *Pipeline) mutateEvent(handler *corev2.Handler, event *corev2.Event) ([]byte, error) {
	// Prepare log entry
	fields := utillogging.EventFields(event, false)
	fields["handler"] = handler.Name

	if handler.Mutator == "" {
		eventData, err := p.jsonMutator(event)

		if err != nil {
			logger.WithFields(fields).WithError(err).Error("failed to mutate event")
			return nil, err
		}

		return eventData, nil
	}

	if handler.Mutator == "only_check_output" {
		if event.HasCheck() {
			eventData := p.onlyCheckOutputMutator(event)
			return eventData, nil
		}
	}

	ctx := context.WithValue(context.Background(), corev2.NamespaceKey, event.Entity.Namespace)
	fields["mutator"] = handler.Mutator

	tctx, cancel := context.WithTimeout(ctx, p.storeTimeout)
	defer cancel()
	mutator, err := p.store.GetMutatorByName(tctx, handler.Mutator)
	if err != nil {
		// Warning: do not wrap this error
		logger.WithFields(fields).WithError(err).Error("failed to retrieve mutator")
		return nil, err
	}

	if mutator == nil {
		// Check to see if there is an extension matching the mutator
		tctx, cancel := context.WithTimeout(ctx, p.storeTimeout)
		defer cancel()
		extension, err := p.store.GetExtension(tctx, handler.Mutator)
		if err != nil {
			if err == store.ErrNoExtension {
				return nil, nil
			}
			// Warning: do not wrap this error
			return nil, err
		}
		executor, err := p.extensionExecutor(extension)
		if err != nil {
			return nil, err
		}
		defer func() {
			if err := executor.Close(); err != nil {
				logger.WithError(err).Debug("error closing grpc client conn")
			}
		}()
		eventData, err := executor.MutateEvent(event)
		if err != nil {
			return nil, err
		}
		return eventData, nil
	}

	eventData, err := p.pipeMutator(mutator, event)

	if err != nil {
		logger.WithFields(fields).WithError(err).Error("failed to mutate the event")
		return nil, err
	}

	return eventData, nil
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
func (p *Pipeline) pipeMutator(mutator *corev2.Mutator, event *corev2.Event) ([]byte, error) {
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
	if len(mutator.RuntimeAssets) != 0 {
		logger.WithFields(fields).Debug("fetching assets for mutator")
		// Fetch and install all assets required for handler execution
		matchedAssets := asset.GetAssets(ctx, p.store, mutator.RuntimeAssets)

		assets, err := asset.GetAll(context.TODO(), p.assetGetter, matchedAssets)
		if err != nil {
			logger.WithFields(fields).WithError(err).Error("failed to retrieve assets for mutator")
		} else {
			mutatorExec.Env = environment.MergeEnvironments(os.Environ(), assets.Env(), mutator.EnvVars, secrets)
		}
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

	logger.WithFields(fields).Debug("event pipe mutator executed")

	return []byte(result.Output), nil
}
