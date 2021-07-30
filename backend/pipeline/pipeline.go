package pipeline

import (
	"context"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/backend/licensing"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/rpc"
	"github.com/sensu/sensu-go/types"
)

type FilterProcessor interface {
	CanFilter(context.Context, *corev3.ResourceReference) bool
	Filter(context.Context, *corev3.ResourceReference, *corev2.Event) (bool, error)
}

type MutatorProcessor interface {
	CanMutate(context.Context, *corev3.ResourceReference) bool
	Mutate(context.Context, *corev3.ResourceReference, *corev2.Event) (*corev2.Event, error)
}

type HandlerProcessor interface {
	CanHandle(context.Context, *corev3.ResourceReference) bool
	Handle(context.Context, *corev3.ResourceReference, *corev2.Event) error
}

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
	filterProcessors       []FilterProcessor
	mutatorProcessors      []MutatorProcessor
	handlerProcessors      []HandlerProcessor
}

func (p *Pipeline) AddFilterProcessor(processor FilterProcessor) {
	p.filterProcessors = append(p.filterProcessors, processor)
}

func (p *Pipeline) AddMutatorProcessor(processor MutatorProcessor) {
	p.mutatorProcessors = append(p.mutatorProcessors, processor)
}

func (p *Pipeline) AddHandlerProcessor(processor HandlerProcessor) {
	p.handlerProcessors = append(p.handlerProcessors, processor)
}

// Config holds the configuration for a Pipeline.
type Config struct {
	Store                   store.Store
	AssetGetter             asset.Getter
	BackendEntity           *corev2.Entity
	ExtensionExecutorGetter ExtensionExecutorGetterFunc
	StoreTimeout            time.Duration
	SecretsProviderManager  *secrets.ProviderManager
	LicenseGetter           licensing.Getter
}

// Option is a functional option used to configure Pipelines.
type Option func(*Pipeline)

// New creates a new Pipeline from the provided configuration.
func New(c Config, options ...Option) *Pipeline {
	defaultFilterProcessors := []FilterProcessor{
		&StandardFilterProcessor{
			assetGetter:       c.AssetGetter,
			extensionExecutor: c.ExtensionExecutorGetter,
			store:             c.Store,
			storeTimeout:      c.StoreTimeout,
		},
	}

	defaultMutatorProcessors := []MutatorProcessor{
		&StandardMutatorProcessor{store: c.Store},
	}

	defaultHandlerProcessors := []HandlerProcessor{
		&StandardHandlerProcessor{store: c.Store},
	}

	pipeline := &Pipeline{
		store:                  c.Store,
		assetGetter:            c.AssetGetter,
		backendEntity:          c.BackendEntity,
		extensionExecutor:      c.ExtensionExecutorGetter,
		executor:               command.NewExecutor(),
		storeTimeout:           c.StoreTimeout,
		secretsProviderManager: c.SecretsProviderManager,
		licenseGetter:          c.LicenseGetter,
		filterProcessors:       defaultFilterProcessors,
		mutatorProcessors:      defaultMutatorProcessors,
		handlerProcessors:      defaultHandlerProcessors,
	}
	for _, o := range options {
		o(pipeline)
	}
	return pipeline
}

const (
	// DefaultSocketTimeout specifies the default socket dial
	// timeout in seconds for TCP and UDP handlers.
	DefaultSocketTimeout uint32 = 60
)

// ExtensionExecutorGetterFunc gets an ExtensionExecutor. Used to decouple
// pipelines from gRPC.
type ExtensionExecutorGetterFunc func(*types.Extension) (rpc.ExtensionExecutor, error)
