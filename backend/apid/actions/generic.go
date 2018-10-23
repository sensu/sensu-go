package actions

import (
	"context"

	storev2 "github.com/sensu/sensu-go/storage"
)

// GenericController exposes generic actions available to most resources
type GenericController struct {
	Store storev2.Store
}

// NewGenericController returns new GenericController
func NewGenericController(store storev2.Store) GenericController {
	return GenericController{
		Store: store,
	}
}

// List returns resources available to the viewer filter by given params.
func (c GenericController) List(ctx context.Context, prefix string, objs interface{}) error {
	return c.Store.List(ctx, prefix, objs)
}
