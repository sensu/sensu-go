package pipeline

import (
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/backend/licensing"
	"github.com/sensu/sensu-go/backend/pipeline/legacy"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/command"
)

// Pipeline takes events as inputs, and treats them in various ways according
// to the event's check configuration.
type Pipeline struct {
	store                  store.Store
	assetGetter            asset.Getter
	backendEntity          *corev2.Entity
	extensionExecutor      ExtensionExecutorGetterFunc
	executor               command.Executor
	storeTimeout           time.Duration
	secretsProviderManager *secrets.ProviderManager
	licenseGetter          licensing.Getter
	filters                []Filter
	mutators               []Mutator
	handlers               []Handler
}

func (p *Pipeline) AddFilter(filter Filter) {
	p.filters = append(p.filters, filter)
}

func (p *Pipeline) AddMutator(mutator Mutator) {
	p.mutators = append(p.mutators, mutator)
}

func (p *Pipeline) AddHandler(handler HandlerProcessor) {
	p.handlers = append(p.handlers, handler)
}

// Config holds the configuration for a Pipeline.
type Config struct {
	Store                  store.Store
	AssetGetter            asset.Getter
	BackendEntity          *corev2.Entity
	StoreTimeout           time.Duration
	SecretsProviderManager *secrets.ProviderManager
	LicenseGetter          licensing.Getter
}

// Option is a functional option used to configure Pipelines.
type Option func(*Pipeline)

// New creates a new Pipeline from the provided configuration.
func New(c Config, options ...Option) *Pipeline {
	// default pipeline filters to search through when searching for a pipeline
	// filter that supports a referenced event filter resource.
	defaultFilters := []Filter{
		&legacy.Filter{
			AssetGetter:  c.AssetGetter,
			Store:        c.Store,
			StoreTimeout: c.StoreTimeout,
		},
	}

	// default pipeline mutators to search through when searching for a pipeline
	// mutator that supports a referenced event mutator resource.
	defaultMutators := []Mutator{
		&legacy.Mutator{
			Store:        c.Store,
			StoreTimeout: c.StoreTimeout,
		},
	}

	// default pipeline handlers to search through when searching for a pipeline
	// handler that supports a referenced event handler resource.
	defaultHandlers := []Handler{
		&legacy.Handler{
			AssetGetter:            c.AssetGetter,
			Executor:               command.NewExecutor(),
			LicenseGetter:          c.LicenseGetter,
			SecretsProviderManager: c.SecretsProviderManager,
			Store:                  c.Store,
			StoreTimeout:           c.StoreTimeout,
		},
	}

	pipeline := &Pipeline{
		store:                  c.Store,
		assetGetter:            c.AssetGetter,
		backendEntity:          c.BackendEntity,
		executor:               command.NewExecutor(),
		storeTimeout:           c.StoreTimeout,
		secretsProviderManager: c.SecretsProviderManager,
		licenseGetter:          c.LicenseGetter,
		filters:                defaultFilters,
		mutators:               defaultMutators,
		handlers:               defaultHandlers,
	}
	for _, o := range options {
		o(pipeline)
	}
	return pipeline
}
