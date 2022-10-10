package routers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	"github.com/sensu/sensu-go/backend/store"
)

// APIKeysRouter handles requests for /apikeys.
type APIKeysRouter struct {
	handlers handlers.Handlers
	store    store.Store
}

// NewAPIKeysRouter instantiates new router for controlling apikeys resources.
func NewAPIKeysRouter(store store.Store) *APIKeysRouter {
	return &APIKeysRouter{
		handlers: handlers.Handlers{
			Resource: &corev2.APIKey{},
			Store:    store,
		},
		store: store,
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
	parent.HandleFunc(routes.PathPrefix, r.create).Methods(http.MethodPost)
	routes.Patch(r.handlers.PatchResource)
}

func (r *APIKeysRouter) create(w http.ResponseWriter, req *http.Request) {
	apikey := &corev2.APIKey{}
	if err := UnmarshalBody(req, apikey); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// validate that the user exists
	if user, err := r.store.GetUser(req.Context(), apikey.Username); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else if user == nil {
		http.Error(w, errors.New("user does not exist").Error(), http.StatusBadRequest)
		return
	}

	// set/overwrite the key id and created_at time
	key, err := uuid.NewRandom()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	apikey.Name = key.String()
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
