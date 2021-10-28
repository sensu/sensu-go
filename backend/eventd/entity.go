package eventd

import (
	"context"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	metricspkg "github.com/sensu/sensu-go/metrics"
)

// createProxyEntity creates a proxy entity for the given event if the entity
// does not exist already and returns the entity created
func createProxyEntity(event *corev2.Event, s storev2.Interface) (fErr error) {
	entityName := event.Entity.Name
	namespace := event.Entity.Namespace

	// Override the entity name with proxy_entity_name if it was provided
	if event.HasCheck() && event.Check.ProxyEntityName != "" {
		entityName = event.Check.ProxyEntityName
	} else if event.Entity.EntityClass == corev2.EntityAgentClass {
		return nil
	}

	begin := time.Now()
	defer func() {
		duration := time.Since(begin)
		status := metricspkg.StatusLabelSuccess
		if fErr != nil {
			status = metricspkg.StatusLabelError
		}
		createProxyEntityDuration.
			WithLabelValues(status).
			Observe(float64(duration) / float64(time.Millisecond))
	}()

	// Determine if the entity exists
	//NOTE(ccressent): there is no timeout for this operation?
	entityMeta := corev2.NewObjectMeta(entityName, namespace)

	state := corev3.NewEntityState(namespace, entityName)
	config := corev3.NewEntityConfig(namespace, entityName)

	configReq := storev2.NewResourceRequestFromResource(context.Background(), config)
	stateReq := storev2.NewResourceRequestFromResource(context.Background(), state)

	// Use postgres when available (enterprise only, entity state only)
	stateReq.UsePostgres = true

	var (
		wState, wConfig storev2.Wrapper
		err             error
	)

	wConfig, err = s.Get(configReq)
	if err == nil {
		if err := wConfig.UnwrapInto(config); err != nil {
			return err
		}

		// Since the entity config exists, we fetch its associated state in
		// order to create a fully formed corev2.Entity for the event.
		wState, err = s.Get(stateReq)
		if err != nil {
			return err
		}

		if err := wState.UnwrapInto(state); err != nil {
			return err
		}
	} else if err != nil {
		switch err.(type) {
		case *store.ErrNotFound:
			// If the entity does not exist, create a proxy entity
			if event.Check.ProxyEntityName != "" {
				// Create a brand new entity since we can't rely on the provided
				// entity, which represents the agent's entity
				state.SetMetadata(&entityMeta)
				config.SetMetadata(&entityMeta)
			} else {
				// Use on the provided entity
				config, state = corev3.V2EntityToV3(event.Entity)
			}

			state.Metadata.CreatedBy = event.CreatedBy

			// Wrap and store the new entity's state. We use CreateOrUpdate()
			// because we want to overwrite any existing EntityState that could
			// have been left behind due to a failed operation or failure to
			// clean up old state.
			wState, err := storev2.WrapResource(state)
			if err != nil {
				return err
			}
			if err := s.CreateOrUpdate(stateReq, wState); err != nil {
				return err
			}

			config.EntityClass = corev2.EntityProxyClass
			config.Subscriptions = append(config.Subscriptions, corev2.GetEntitySubscription(entityName))

			// Wrap and store the new entity's configuration. We use
			// CreateIfNotExists() to assert that this EntityConfig is indeed
			// brand new.
			wConfig, err := storev2.WrapResource(config)
			if err != nil {
				return err
			}
			if err := s.CreateIfNotExists(configReq, wConfig); err != nil {
				return err
			}
		default:
			return err
		}
	}

	entity, err := corev3.V3EntityToV2(config, state)
	if err != nil {
		return err
	}

	// Replace the event's entity with the proxy entity
	event.Entity = entity
	return nil
}
