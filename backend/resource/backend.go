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
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
	"github.com/sensu/sensu-go/system"
)

type eventInfo struct {
	status       uint32
	timestampSec int64
}

type BackendResource struct {
	store             storev2.Interface
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

func New(store storev2.Interface, bus messaging.MessageBus) *BackendResource {
	return &BackendResource{
		store:             store,
		bus:               bus,
		lastEvents:        map[string]*eventInfo{},
		repeatIntervalSec: int64(30),
	}
}

func (br *BackendResource) EnsureBackendResources(ctx context.Context) error {
	nsStore := br.store.NamespaceStore()
	_, err := nsStore.Get(ctx, systemNamespaceName)
	switch err.(type) {
	case *store.ErrNotFound:
		err = nsStore.CreateIfNotExists(ctx, corev3.NewNamespace(systemNamespaceName))
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
	backendEntityConfigReq := storev2.NewResourceRequestFromResource(backendEntityConfig)
	wrappedBackendEntityConfig, err := wrap.Resource(backendEntityConfig)
	if err != nil {
		return err
	}
	if err := br.store.CreateOrUpdate(ctx, backendEntityConfigReq, wrappedBackendEntityConfig); err != nil {
		return err
	}

	backendEntityStateReq := storev2.NewResourceRequestFromResource(backendEntityState)
	wrappedBackendEntityState, err := wrap.Resource(backendEntityState)
	if err != nil {
		return err
	}
	if err := br.store.CreateOrUpdate(ctx, backendEntityStateReq, wrappedBackendEntityState); err != nil {
		return err
	}

	br.backendEntity = backendEntity

	return nil
}

func (br *BackendResource) GenerateBackendEvent(component string, status uint32, output string) error {
	if br.backendEntity == nil {
		return errors.New("backend entity doesn't exist")
	}

	now := time.Now().Unix()
	if lastEvent, ok := br.lastEvents[component]; ok {
		if lastEvent.status == status && now-lastEvent.timestampSec < br.repeatIntervalSec {
			return nil
		}
	}

	id := uuid.New()
	event := &corev2.Event{
		Timestamp: now,
		Entity:    br.backendEntity,
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
	err := br.bus.Publish(messaging.TopicEventRaw, event)
	if err == nil {
		br.lastEvents[component] = &eventInfo{
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
