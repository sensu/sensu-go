package actions

import (
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	"golang.org/x/net/context"
)

// TessenController exposes actions which a viewer can perform
type TessenController struct {
	store store.TessenConfigStore
	bus   messaging.MessageBus
}

// NewTessenController returns a new TessenController
func NewTessenController(store store.TessenConfigStore, bus messaging.MessageBus) TessenController {
	return TessenController{
		store: store,
		bus:   bus,
	}
}

// CreateOrUpdate creates or updates the tessen configuration
func (c TessenController) CreateOrUpdate(ctx context.Context, config *corev2.TessenConfig) error {
	if err := c.store.CreateOrUpdateTessenConfig(ctx, config); err != nil {
		switch err := err.(type) {
		case *store.ErrNotValid:
			return NewErrorf(InvalidArgument)
		default:
			return NewError(InternalErr, err)
		}
	}

	// Publish to Tessend
	if err := c.bus.Publish(messaging.TopicTessen, config); err != nil {
		return NewError(InternalErr, err)
	}

	return nil
}

// Get gets the tessen configuration
func (c TessenController) Get(ctx context.Context) (*corev2.TessenConfig, error) {
	config, err := c.store.GetTessenConfig(ctx)
	if err != nil {
		switch err := err.(type) {
		case *store.ErrNotFound:
			return nil, NewErrorf(NotFound)
		default:
			return nil, NewError(InternalErr, err)
		}
	}

	return config, nil
}
