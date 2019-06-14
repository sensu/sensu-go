package actions

import (
	"github.com/sensu/sensu-go/backend/store"
	"golang.org/x/net/context"
)

// ClusterIDController exposes actions which a viewer can perform
type ClusterIDController struct {
	store store.ClusterIDStore
}

// NewClusterIDController returns a new ClusterIDController
func NewClusterIDController(store store.ClusterIDStore) ClusterIDController {
	return ClusterIDController{
		store: store,
	}
}

// Get gets the sensu cluster id
func (c ClusterIDController) Get(ctx context.Context) (string, error) {
	id, err := c.store.GetClusterID(ctx)
	if err != nil {
		switch err := err.(type) {
		case *store.ErrNotFound:
			return "", NewErrorf(NotFound)
		default:
			return "", NewError(InternalErr, err)
		}
	}

	return id, nil
}
