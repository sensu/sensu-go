package routers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	"github.com/sensu/sensu-go/backend/store"
)

// APIKeysRouter handles requests for /apikeys.
type APIKeysRouter struct {
	handlers handlers.Handlers
}

// NewAPIKeysRouter instantiates new router for controlling apikeys resources.
func NewAPIKeysRouter(store store.ResourceStore) *APIKeysRouter {
	return &APIKeysRouter{
		handlers: handlers.Handlers{
			Resource: &corev2.APIKey{},
			Store:    store,
		},
	}
}

// Mount the APIKeysRouter to a parent Router.
func (r *APIKeysRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/{resource:apikeys}",
	}

	routes.Del(r.handlers.DeleteResource)
	routes.Get(r.handlers.GetResource)
	routes.List(r.handlers.ListResources, corev2.APIKeyFields)
	routes.Patch(r.update)
	parent.HandleFunc(routes.PathPrefix, r.create).Methods(http.MethodPost)
}

func (r *APIKeysRouter) create(w http.ResponseWriter, req *http.Request) {
	apikey := &corev2.APIKey{}
	if err := UnmarshalBody(req, apikey); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// set/overwrite the key id and created_at time
	apikey.Key = uuid.New().String()
	apikey.CreatedAt = time.Now().Unix()
	newBytes, err := json.Marshal(apikey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Body = ioutil.NopCloser(bytes.NewBuffer(newBytes))

	_, err = r.handlers.CreateResource(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// set the relative location header
	w.Header().Set("Location", fmt.Sprintf("%s/%s", req.URL.String(), apikey.Name))
	w.WriteHeader(http.StatusCreated)
}

func (r *APIKeysRouter) update(req *http.Request) (interface{}, error) {
	stored := &corev2.APIKey{}
	params := mux.Vars(req)
	id, err := url.PathUnescape(params["id"])
	if err != nil {
		return nil, err
	}
	if err := r.handlers.Store.GetResource(req.Context(), id, stored); err != nil {
		return nil, actions.NewError(actions.NotFound, err)
	}

	apikey := &corev2.APIKey{}
	if err := UnmarshalBody(req, apikey); err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}

	// only allow groups to be updated
	stored.Groups = apikey.Groups
	return stored, r.handlers.Store.CreateOrUpdateResource(req.Context(), stored)
}
