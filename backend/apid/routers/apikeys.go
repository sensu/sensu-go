package routers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/echlebek/pet"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	"github.com/sensu/sensu-go/backend/authentication/bcrypt"
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
	routes.List(handlers.ListResources, corev3.APIKeyFields)
	parent.HandleFunc(routes.PathPrefix, r.create).Methods(http.MethodPost, http.MethodPut)
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
		// validate that the user the API key pertains to exists
		if _, ok := err.(*store.ErrNotFound); ok {
			http.Error(w, errors.New("user does not exist").Error(), http.StatusBadRequest)
			return
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	response := corev2.APIKeyResponse{}
	if len(bytes.TrimSpace(apikey.Hash)) == 0 {
		// If Hash is not specified by the client, we generate a new one for them,
		// and return the secret key in the response body. Otherwise, {} is returned.
		secretKey, err := uuid.NewRandom()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		hash, err := bcrypt.HashPassword(secretKey.String())
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		apikey.Hash = []byte(hash)
		response.Key = secretKey.String()
	}
	apikey.CreatedAt = time.Now().Unix()
	if strings.TrimSpace(apikey.Name) == "" {
		// If the API key is not named, generate a pet name for the API key.
		apikey.Name = pet.Generate(3, "")
	}

MARSHAL:

	newBytes, err := json.Marshal(apikey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	req.Body = ioutil.NopCloser(bytes.NewBuffer(newBytes))

	handlers := handlers.NewHandlers[*corev2.APIKey](r.store)

	createFunc := handlers.CreateResource
	if req.Method == http.MethodPut {
		createFunc = handlers.CreateOrUpdateResource
	}

	_, err = createFunc(req)
	if err != nil {
		if actionErr, ok := err.(actions.Error); ok {
			// small chance of pet name collision, this should take care of it
			if actionErr.Code == actions.AlreadyExistsErr {
				apikey.Name = pet.Generate(3, "")
				goto MARSHAL
			}
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response.Name = apikey.Name

	// set the relative location header
	w.Header().Set("Location", fmt.Sprintf("%s/%s", req.URL.String(), apikey.Name))
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Error(err)
	}
}
