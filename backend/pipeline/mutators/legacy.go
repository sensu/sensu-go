package legacy

import (
	"context"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/backend/store"
	utillogging "github.com/sensu/sensu-go/util/logging"
)

type Mutator struct {
	Store        store.Store
	StoreTimeout time.Duration
}

func (m *Mutator) CanMutate(ctx context.Context, ref *corev2.ResourceReference) bool {
	if ref.APIVersion == "core/v2" && ref.Type == "Mutator" {
		return true
	}
	return false
}

// Mutate mutates (transforms) a Sensu event into a serialized format
// (byte slice) to be provided to a Sensu event handler.
func (m *Mutator) Mutate(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event) (*corev2.Event, error) {
	// Prepare log entry
	// TODO: add pipeline & pipeline workflow names to fields
	fields := utillogging.EventFields(event, false)

	if handler.Mutator == "" {
		eventData, err := p.jsonMutator(event)
		if err != nil {
			logger.WithFields(fields).WithError(err).Error("failed to mutate event")
			return nil, err
		}

		return eventData, nil
	}

	if ref.Name == "only_check_output" {
		if event.HasCheck() {
			eventData := s.onlyCheckOutputMutator(event)
			return eventData, nil
		}
	}

	ctx = context.WithValue(ctx, corev2.NamespaceKey, event.Entity.Namespace)

	tctx, cancel := context.WithTimeout(ctx, m.storeTimeout)
	defer cancel()
	mutator, err := m.store.GetMutatorByName(tctx, ref.Name)
	if err != nil {
		// Warning: do not wrap this error
		logger.WithFields(fields).WithError(err).Error("failed to retrieve mutator")
		return nil, err
	}

	if mutator == nil {
		// TODO: handle this
	}

	var eventData []byte

	var assets asset.RuntimeAssetSet
	if len(mutator.RuntimeAssets) > 0 {
		logger.WithFields(fields).Debug("fetching assets for mutator")

		// Fetch and install all assets required for handler execution
		matchedAssets := asset.GetAssets(ctx, p.store, mutator.RuntimeAssets)

		var err error
		assets, err = asset.GetAll(ctx, p.assetGetter, matchedAssets)
		if err != nil {
			logger.WithFields(fields).WithError(err).Error("failed to retrieve assets for mutator")
		}
	}

	if mutator.Type == "" || mutator.Type == corev2.PipeMutator {
		eventData, err = p.pipeMutator(mutator, event, assets)
	} else if mutator.Type == corev2.JavascriptMutator {
		eventData, err = p.javascriptMutator(mutator, event, assets)
	}

	if err != nil {
		logger.WithFields(fields).WithError(err).Error("failed to mutate the event")
		return nil, err
	}

	return eventData, nil
	return nil, nil
}
