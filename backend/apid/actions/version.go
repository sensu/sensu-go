package actions

import (
	"github.com/coreos/etcd/client"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"golang.org/x/net/context"
)

// VersionController exposes actions which a viewer can perform
type VersionController struct {
	store  store.VersionStore
	client client.Client
}

// NewVersionController returns a new VersionController
func NewVersionController(store store.VersionStore, client client.Client) VersionController {
	return VersionController{
		store:  store,
		client: client,
	}
}

// GetVersion returns version information
func (v VersionController) GetVersion(ctx context.Context) (*corev2.Version, error) {
	return v.store.GetVersion(ctx, v.client)
}
