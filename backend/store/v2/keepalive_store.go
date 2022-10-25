package v2

import (
	"context"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/store"
)

// NOTE(ccressent): Given how we are bound to rework keepalives "soon", this is
// just a "wrapper" that calls the storev1 methods. Subject to a better name if
// anyone has one.
type LegacyKeepaliveStore struct {
	store store.KeepaliveStore
}

func NewLegacyKeepaliveStore(store store.KeepaliveStore) *LegacyKeepaliveStore {
	return &LegacyKeepaliveStore{
		store: store,
	}
}

func (s *LegacyKeepaliveStore) DeleteFailingKeepalive(ctx context.Context, entityConfig *corev3.EntityConfig) error {
	// Transform *corev3.EntityConfig into a usable *corev2.Entity
	emptyEntityState := corev3.NewEntityState(entityConfig.Metadata.Namespace, entityConfig.Metadata.Name)
	v2Entity, err := corev3.V3EntityToV2(entityConfig, emptyEntityState)
	if err != nil {
		return err
	}

	return s.store.DeleteFailingKeepalive(ctx, v2Entity)
}

func (s *LegacyKeepaliveStore) GetFailingKeepalives(ctx context.Context) ([]*corev2.KeepaliveRecord, error) {
	return s.store.GetFailingKeepalives(ctx)
}

func (s *LegacyKeepaliveStore) UpdateFailingKeepalive(ctx context.Context, entityConfig *corev3.EntityConfig, expiration int64) error {
	// Transform *corev3.EntityConfig into a usable *corev2.Entity
	emptyEntityState := corev3.NewEntityState(entityConfig.Metadata.Namespace, entityConfig.Metadata.Name)
	v2Entity, err := corev3.V3EntityToV2(entityConfig, emptyEntityState)
	if err != nil {
		return err
	}

	return s.store.UpdateFailingKeepalive(ctx, v2Entity, expiration)
}
