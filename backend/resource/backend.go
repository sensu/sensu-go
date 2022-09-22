package resource

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/messaging"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/system"
)

type eventInfo struct {
	status       uint32
	timestampSec int64
}

type BackendResource struct {
	namespaceStore    storev2.NamespaceStore
	entityConfigStore storev2.EntityConfigStore
	entityStateStore  storev2.EntityStateStore
	bus               messaging.MessageBus
	backendEntity     *corev2.Entity
	lastEvents        map[string]*eventInfo
	repeatIntervalSec int64
}

const (
	// ComponentSecrets represents the secrets component of the Sensu backend
	ComponentSecrets = "secrets"

	// The default Sensu system namespace
	systemNamespaceName = "sensu-system"
)

func New(ns storev2.NamespaceStore, entc storev2.EntityConfigStore, ents storev2.EntityStateStore, bus messaging.MessageBus) *BackendResource {
	return &BackendResource{
		namespaceStore:    ns,
		entityConfigStore: entc,
		entityStateStore:  ents,
		bus:               bus,
		lastEvents:        map[string]*eventInfo{},
		repeatIntervalSec: int64(30),
	}
}

func (b *BackendResource) EnsureBackendResources(ctx context.Context) error {
	_, err := b.namespaceStore.Get(ctx, systemNamespaceName)
	switch err.(type) {
	case *store.ErrNotFound:
		err = b.namespaceStore.CreateIfNotExists(ctx, corev3.NewNamespace(systemNamespaceName))
		if err != nil {
			return err
		}
	default:
		return err
	}

	backendEntity, err := getEntity()
	if err != nil {
		return err
	}

	backendEntityConfig, backendEntityState := corev3.V2EntityToV3(backendEntity)
	if err != nil {
		return err
	}

	if err := b.entityConfigStore.CreateOrUpdate(ctx, backendEntityConfig); err != nil {
		return err
	}

	if err := b.entityStateStore.CreateOrUpdate(ctx, backendEntityState); err != nil {
		return err
	}

	b.backendEntity = backendEntity

	return nil
}

func (b *BackendResource) GenerateBackendEvent(component string, status uint32, output string) error {
	if b.backendEntity == nil {
		return errors.New("backend entity doesn't exist")
	}

	now := time.Now().Unix()
	if lastEvent, ok := b.lastEvents[component]; ok {
		if lastEvent.status == status && now-lastEvent.timestampSec < b.repeatIntervalSec {
			return nil
		}
	}

	id := uuid.New()
	event := &corev2.Event{
		Timestamp: now,
		Entity:    b.backendEntity,
		Check: &corev2.Check{
			ObjectMeta: corev2.NewObjectMeta(component, systemNamespaceName),
			Issued:     now,
			Executed:   now,
			Output:     output,
			Status:     status,
		},
		ObjectMeta: corev2.NewObjectMeta("", systemNamespaceName),
		ID:         id[:],
	}
	err := b.bus.Publish(messaging.TopicEventRaw, event)
	if err == nil {
		b.lastEvents[component] = &eventInfo{
			status:       status,
			timestampSec: now,
		}
	}

	return err
}

func getEntity() (*corev2.Entity, error) {
	systemInfo, err := system.Info()
	if err != nil {
		return nil, err
	}
	meta := corev2.NewObjectMeta(systemInfo.Hostname, systemNamespaceName)
	return &corev2.Entity{
		EntityClass: "backend",
		ObjectMeta:  meta,
		System:      systemInfo,
	}, nil
}
