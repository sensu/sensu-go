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
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

// APIKeysRouter handles requests for /apikeys.
type APIKeysRouter struct {
	store storev2.Interface
}

// NewAPIKeysRouter instantiates new router for controlling apikeys resources.
func NewAPIKeysRouter(store storev2.Interface) *APIKeysRouter {
	return &APIKeysRouter{
		store: store,
	}
}

// Mount the APIKeysRouter to a parent Router.
func (r *APIKeysRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/{resource:apikeys}",
	}

	handlers := handlers.NewHandlers[*corev2.APIKey](r.store)

	routes.Del(handlers.DeleteResource)
	routes.Get(handlers.GetResource)
	routes.List(handlers.ListResources, corev2.APIKeyFields)
	parent.HandleFunc(routes.PathPrefix, r.create).Methods(http.MethodPost)
	routes.Patch(handlers.PatchResource)
}

func (r *APIKeysRouter) create(w http.ResponseWriter, req *http.Request) {
	apikey := &corev2.APIKey{}
	if err := UnmarshalBody(req, apikey); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// validate that the user exists
	user := &corev2.User{Username: apikey.Username}
	storeReq := storev2.NewResourceRequestFromResource(user)
	if _, err := r.store.GetConfigStore().Get(req.Context(), storeReq); err != nil {
		if _, ok := err.(*store.ErrNotFound); ok {
			http.Error(w, errors.New("user does not exist").Error(), http.StatusBadRequest)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
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

	handlers := handlers.NewHandlers[*corev2.APIKey](r.store)

	_, err = handlers.CreateResource(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// set the relative location header
	w.Header().Set("Location", fmt.Sprintf("%s/%s", req.URL.String(), apikey.Name))
	w.WriteHeader(http.StatusCreated)
}
