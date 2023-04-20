package filter

import (
	"context"
	"errors"
	"time"

	"github.com/robertkrimen/otto"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/js"
	"github.com/sensu/sensu-go/dynamic"
)

const (
	pipelineRoleName = "system:pipeline"

	// LegacyAdapterName is the name of the filter adapter.
	LegacyAdapterName = "LegacyAdapter"
)

var (
	builtInFilterNames = []string{
		"is_incident",
		"has_metrics",
		"not_silenced",
	}

	errCouldNotRetrieveFilter = errors.New("could not retrieve filter")

	// PipelineFilterFuncs gets patched by enterprise sensu-go
	PipelineFilterFuncs map[string]interface{}
)

// LegacyAdapter is a filter adapter that supports the legacy
// core.v2/EventFilter type.
type LegacyAdapter struct {
	AssetGetter  asset.Getter
	Store        storev2.Interface
	StoreTimeout time.Duration
}

// Name returns the name of the filter adapter.
func (l *LegacyAdapter) Name() string {
	return LegacyAdapterName
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
	fields := event.LogFields(false)
	fields["pipeline"] = corev2.ContextPipeline(ctx)
	fields["pipeline_workflow"] = corev2.ContextPipelineWorkflow(ctx)

	// Retrieve the filter from the store with its name
	fstore := storev2.Of[*corev2.EventFilter](l.Store)
	ctx = context.WithValue(ctx, corev2.NamespaceKey, event.Entity.Namespace)
	tctx, cancel := context.WithTimeout(ctx, l.StoreTimeout)

	filter, err := fstore.Get(tctx, storev2.ID{Namespace: event.Entity.Namespace, Name: ref.Name})
	cancel()
	if err != nil {
		logger.WithFields(fields).WithError(err).Warning(errCouldNotRetrieveFilter.Error())
		return false, err
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
	fields["action"] = filter.Action
	fields["filter"] = filter.Name
	fields["assets"] = filter.RuntimeAssets
	fields["pipeline"] = corev2.ContextPipeline(ctx)
	fields["pipeline_workflow"] = corev2.ContextPipelineWorkflow(ctx)

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

		if filter.Action == corev2.EventFilterActionDeny && inWindows {
			logger.WithFields(fields).Debug("denying event inside of filtering window")
			return true
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

	switch filter.Action {
	// Inclusive "Allow" filters let events through when the AND'd combination
	// of their expressions is true. That is, events that match all of the
	// expressions are not filtered.
	case corev2.EventFilterActionAllow:

		for _, expression := range filter.Expressions {
			match, err := env.Eval(ctx, expression)
			if err != nil {
				logger.WithFields(fields).WithError(err).Error("error evaluating javascript event filter")
				continue
			}

			// One of the expressions did not match, filter the event
			if !match {
				logger.WithFields(fields).Debug("denying event that does not match filter")
				return true
			}
		}

		// All the expressions matched, do not filter the event
		logger.WithFields(fields).Debug("allowing event that matches filter")
		return false

	// Exclusive "Deny" filters let events through when the OR'd combination of
	// their expressions is true. That is, events that match one or more of the
	// expressions are filtered.
	case corev2.EventFilterActionDeny:

		for _, expression := range filter.Expressions {
			match, err := env.Eval(ctx, expression)
			if err != nil {
				logger.WithFields(fields).WithError(err).Error("error evaluating javascript event filter")
				continue
			}

			// One of the expressions matched, filter the event
			if match {
				logger.WithFields(fields).Debug("denying event that matches filter")
				return true
			}
		}

		// None of the expressions matched, do not filter the event
		logger.WithFields(fields).Debug("allowing event that does not match filter")
		return false

	default:
		logger.WithFields(fields).Error("unrecognized filter action")
		return false
	}
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
