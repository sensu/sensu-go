package actions

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sirupsen/logrus"
)

type EntityDeleter struct {
	EntityStore store.EntityStore
	EventStore  store.EventStore
}

func (d EntityDeleter) Delete(req *http.Request) (corev3.Resource, error) {
	params := mux.Vars(req)
	entityName, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, NewError(InvalidArgument, err)
	}

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

	result, err := d.EntityStore.GetEntityByName(req.Context(), entityName)
	if err != nil {
		return nil, NewError(InternalErr, err)
	}

	if result == nil {
		return nil, NewErrorf(NotFound)
	}

	return nil, d.EntityStore.DeleteEntityByName(req.Context(), entityName)
}
