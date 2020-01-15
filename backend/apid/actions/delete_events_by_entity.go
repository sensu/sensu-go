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
	Store store.Store
}

func (d EntityDeleter) Delete(req *http.Request) (interface{}, error) {
	params := mux.Vars(req)
	entityName, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, NewError(InvalidArgument, err)
	}

	events, err := d.Store.GetEventsByEntity(req.Context(), entityName, &store.SelectionPredicate{})
	if err != nil {
		return nil, fmt.Errorf("error fetching events for entity: %s", err)
	}

	for _, event := range events {
		if !event.HasCheck() {
			// improbable
			continue
		}
		err := d.Store.DeleteEventByEntityCheck(req.Context(), entityName, event.Check.Name)
		if err != nil {
			logger := logger.WithFields(logrus.Fields{
				"entity":    entityName,
				"check":     event.Check.Name,
				"namespace": event.Namespace})
			logger.WithError(err).Error("error deleting event from entity")
			continue
		}
	}

	result, err := d.Store.GetEntityByName(req.Context(), entityName)
	if err != nil {
		return nil, NewError(InternalErr, err)
	}

	if result == nil {
		return nil, NewErrorf(NotFound)
	}

	return nil, d.Store.DeleteEntityByName(req.Context(), entityName)
}
