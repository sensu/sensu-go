package actions

import (
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/tessen"
	"golang.org/x/net/context"
)

// TessenController exposes actions which a viewer can perform
type TessenController struct {
	store store.TessenConfigStore
}

// NewTessenController returns a new TessenController
func NewTessenController(store store.TessenConfigStore) TessenController {
	return TessenController{
		store: store,
	}
}

// CreateOrUpdate creates or updates the tessen configuration
func (c TessenController) CreateOrUpdate(ctx context.Context, config *tessen.Config) error {
	if err := c.store.CreateOrUpdateTessenConfig(ctx, config); err != nil {
		switch err := err.(type) {
		case *store.ErrNotValid:
			return NewErrorf(InvalidArgument)
		default:
			return NewError(InternalErr, err)
		}
	}

	return nil
}

// Get gets the tessen configuration
func (c TessenController) Get(ctx context.Context) (*tessen.Config, error) {
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
