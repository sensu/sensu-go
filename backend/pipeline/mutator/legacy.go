package mutator

import (
	"context"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/command"
	utillogging "github.com/sensu/sensu-go/util/logging"
)

var (
	builtInMutatorNames = []string{
		"json",
		"only_check_output",
	}
)

// Legacy is
type Legacy struct {
	AssetGetter            asset.Getter
	Executor               command.Executor
	SecretsProviderManager *secrets.ProviderManager
	Store                  store.Store
	StoreTimeout           time.Duration
}

// Name returns the name of the pipeline mutator.
func (l *Legacy) Name() string {
	return "Legacy"
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
		pipeMutator := &Pipe{
			AssetGetter:            l.AssetGetter,
			Executor:               l.Executor,
			SecretsProviderManager: l.SecretsProviderManager,
			Store:                  l.Store,
			StoreTimeout:           l.StoreTimeout,
		}
		eventData, err = pipeMutator.run(ctx, mutator, event, assets)
	} else if mutator.Type == corev2.JavascriptMutator {
		javascriptMutator := Javascript{
			AssetGetter:  l.AssetGetter,
			Store:        l.Store,
			StoreTimeout: l.StoreTimeout,
		}
		eventData, err = javascriptMutator.run(ctx, mutator, event, assets)
	}

	if err != nil {
		logger.WithFields(fields).WithError(err).Error("failed to mutate the event")
		return nil, err
	}

	return eventData, nil
}
