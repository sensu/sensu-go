package filter

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/robertkrimen/otto"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/js"
	"github.com/sensu/sensu-go/types/dynamic"
)

const pipelineRoleName = "system:pipeline"

var (
	builtInFilterNames = []string{
		"is_incident",
		"has_metrics",
		"not_silenced",
	}

	getFilterErr = errors.New("could not retrieve filter")

	// PipelineFilterFuncs gets patched by enterprise sensu-go
	PipelineFilterFuncs map[string]interface{}
)

// LegacyAdapter is a filter adapter that supports the legacy
// core.v2/EventFilter type.
type LegacyAdapter struct {
	AssetGetter  asset.Getter
	Store        store.Store
	StoreTimeout time.Duration
}

// Name returns the name of the filter adapter.
func (l *LegacyAdapter) Name() string {
	return "LegacyAdapter"
}

// CanFilter determines whether LegacyAdapter can filter the resource being
// referenced.
func (l *LegacyAdapter) CanFilter(ref *corev2.ResourceReference) bool {
	if ref.APIVersion == "core/v2" && ref.Type == "EventFilter" {
		for _, name := range builtInFilterNames {
			if ref.Name == name {
				return false
			}
		}
		return true
	}
	return false
}

// Filter filters a Sensu event, determining if it will continue through the
// Sensu pipeline. It returns whether or not the event was filtered and if any
// error was encountered.
func (l *LegacyAdapter) Filter(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event) (bool, error) {
	// Prepare log entry
	// TODO: add pipeline & pipeline workflow names to fields
	fields := event.LogFields(false)

	// Retrieve the filter from the store with its name
	ctx = corev2.SetContextFromResource(context.Background(), event.Entity)
	tctx, cancel := context.WithTimeout(ctx, l.StoreTimeout)

	filter, err := l.Store.GetEventFilterByName(tctx, ref.Name)
	cancel()
	if err != nil {
		logger.WithFields(fields).WithError(err).Warning(getFilterErr.Error())
		return false, err
	}
	if filter == nil {
		logger.WithFields(fields).WithError(err).Warning(getFilterErr.Error())
		return false, fmt.Errorf(getFilterErr.Error())
	}

	// Execute the filter, evaluating each of its
	// expressions against the event. The event is rejected
	// if the product of all expressions is true.
	ctx = corev2.SetContextFromResource(ctx, filter)
	matchedAssets := asset.GetAssets(ctx, l.Store, filter.RuntimeAssets)
	assets, err := asset.GetAll(ctx, l.AssetGetter, matchedAssets)
	if err != nil {
		logger.WithFields(fields).WithError(err).Error("failed to retrieve assets for filter")
		if _, ok := err.(*store.ErrInternal); ok {
			// Fatal error
			return false, err
		}
	}
	filtered := evaluateEventFilter(ctx, event, filter, assets)
	if filtered {
		logger.WithFields(fields).Debug("denying event with custom filter")
		return true, nil
	}

	logger.WithFields(fields).Debug("allowing event")
	return false, nil
}

// Returns true if the event should be filtered/denied.
func evaluateEventFilter(ctx context.Context, event *corev2.Event, filter *corev2.EventFilter, assets asset.RuntimeAssetSet) bool {
	// Redact the entity to avoid leaking sensitive information
	event.Entity = event.Entity.GetRedactedEntity()

	fields := event.LogFields(false)
	fields["filter"] = filter.Name
	fields["assets"] = filter.RuntimeAssets

	if filter.When != nil {
		inWindows, err := filter.When.InWindows(time.Now().UTC())
		if err != nil {
			logger.WithFields(fields).WithError(err).
				Error("denying event - unable to determine if time is in specified window")
			return false
		}

		if filter.Action == corev2.EventFilterActionAllow && !inWindows {
			logger.WithFields(fields).Debug("denying event outside of filtering window")
			return true
		}

		if filter.Action == corev2.EventFilterActionDeny && !inWindows {
			logger.WithFields(fields).Debug("allowing event outside of filtering window")
			return false
		}
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

	synth := dynamic.Synthesize(event)
	env := FilterExecutionEnvironment{
		Event:  synth,
		Assets: assets,
		Funcs:  PipelineFilterFuncs,
	}

	for _, expression := range filter.Expressions {
		match, err := env.Eval(ctx, expression)
		if err != nil {
			logger.WithFields(fields).WithError(err).Error("error evaluating javascript event filter")
			continue
		}

		// Allow - One of the expressions did not match, filter the event
		if filter.Action == corev2.EventFilterActionAllow && !match {
			logger.WithFields(fields).Debug("denying event that does not match filter")
			return true
		}

		// Deny - One of the expressions did not match, do not filter the event
		if filter.Action == corev2.EventFilterActionDeny && !match {
			logger.WithFields(fields).Debug("allowing event that does not match filter")
			return false
		}
	}

	// Allow - All of the expressions matched, do not filter the event
	if filter.Action == corev2.EventFilterActionAllow {
		logger.WithFields(fields).Debug("allowing event that matches filter")
		return false
	}

	// Deny - All of the expressions matched, filter the event
	if filter.Action == corev2.EventFilterActionDeny {
		logger.WithFields(fields).Debug("denying event that matches filter")
		return true
	}

	// Something weird happened, let's not filter the event and log a warning message
	logger.WithFields(fields).
		Warn("not filtering event due to unhandled case")

	return false
}

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
