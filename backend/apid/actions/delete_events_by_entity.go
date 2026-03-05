package actions

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sirupsen/logrus"
)

type EntityDeleter struct {
	EntityStore   store.EntityStore
	EventStore    store.EventStore
	SilencedStore store.SilencedStore
}

func (d EntityDeleter) Delete(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	entityName, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, NewError(InvalidArgument, err)
	}

	// Delete all events associated with the entity
	events, err := d.EventStore.GetEventsByEntity(req.Context(), entityName, &store.SelectionPredicate{})
	if err != nil {
		return nil, fmt.Errorf("error fetching events for entity: %s", err)
	}

	for _, event := range events {
		if !event.HasCheck() {
			// improbable
			continue
		}
		err := d.EventStore.DeleteEventByEntityCheck(req.Context(), entityName, event.Check.Name)
		if err != nil {
			logger := logger.WithFields(logrus.Fields{
				"entity":    entityName,
				"check":     event.Check.Name,
				"namespace": event.Namespace})
			logger.WithError(err).Error("error deleting event from entity")
			continue
		}
	}

	// Delete entity-specific silenced entries (e.g., entity:web-01:*)
	if d.SilencedStore != nil {
		entitySubscription := fmt.Sprintf("entity:%s", entityName)
		silencedEntries, err := d.SilencedStore.GetSilencedEntriesBySubscription(req.Context(), entitySubscription)
		if err != nil {
			// Log the error but continue with entity deletion
			logger.WithError(err).WithFields(logrus.Fields{
				"entity":       entityName,
				"subscription": entitySubscription,
			}).Warn("error fetching silenced entries for entity")
		} else if len(silencedEntries) > 0 {
			silencedNames := make([]string, 0, len(silencedEntries))
			for _, entry := range silencedEntries {
				silencedNames = append(silencedNames, entry.Name)
			}

			err = d.SilencedStore.DeleteSilencedEntryByName(req.Context(), silencedNames...)
			if err != nil {
				// Log the error but continue with entity deletion
				logger.WithError(err).WithFields(logrus.Fields{
					"entity":   entityName,
					"silences": silencedNames,
				}).Warn("error deleting silenced entries for entity")
			} else {
				logger.WithFields(logrus.Fields{
					"entity":   entityName,
					"silences": silencedNames,
					"count":    len(silencedNames),
				}).Info("deleted entity-specific silenced entries")
			}
		}
	}

	// Verify the entity exists before attempting deletion
	result, err := d.EntityStore.GetEntityByName(req.Context(), entityName)
	if err != nil {
		return nil, NewError(InternalErr, err)
	}

	if result == nil {
		return nil, NewErrorf(NotFound)
	}

	// Delete the entity
	return nil, d.EntityStore.DeleteEntityByName(req.Context(), entityName)
}
