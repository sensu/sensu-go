package controllers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/useractions"
)

type errorBody struct {
	Error string `json:"error"`
	Code  uint32 `json:"code"`
}

func respondWith(w http.ResponseWriter, resources interface{}) {
	// Set content-type to JSON
	w.Header().Set("Content-Type", "application/json")

	// If not resource(s) are present return a 204 response code
	if resources == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	bytes, err := json.Marshal(resources)
	if err != nil {
		writeError(w, err)
		return
	}

	if _, err := w.Write(bytes); err != nil {
		logger.WithError(err).Error("failed to write response")
		writeError(w, err)
	}
}

func writeError(w http.ResponseWriter, err error) {
	const fallback = `{"error": "failed to marshal error message"}`

	errBody := errorBody{}
	st := http.StatusInternalServerError

	// Wrap message in standard errorBody
	actionErr, ok := err.(useractions.Error)
	if ok {
		errBody.Error = actionErr.Message
		errBody.Code = uint32(actionErr.Code)
		st = HTTPStatusFromCode(actionErr.Code)
	} else {
		errBody.Error = err.Error()
	}

	// Prevent browser from doing mime-sniffing
	w.Header().Set("X-Content-Type-Options", "nosniff")

	// Marshall error message to JSON
	errJSON, err := json.Marshal(errBody)
	if err != nil {
		logEntry := logger.WithField("errBody", errBody).WithError(err)
		logEntry.Error("failed to serialize error body")
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := io.WriteString(w, fallback); err != nil {
			logEntry = logger.WithError(err)
			logEntry.Error("failed to write response")
		}
		return
	}

	// Write JSON
	w.WriteHeader(st)
	fmt.Println(w, errJSON)
}

// HTTPStatusFromCode returns http status code for given user action err code
func HTTPStatusFromCode(code useractions.ErrCode) int {
	switch code {
	case useractions.InternalErr:
		return http.StatusInternalServerError
	case useractions.InvalidArgument:
		return http.StatusBadRequest
	case useractions.NotFound:
		return http.StatusNotFound
	case useractions.AlreadyExistsErr:
		return http.StatusConflict
	case useractions.PermissionDenied:
		return http.StatusUnauthorized
	case useractions.Unauthenticated:
		return http.StatusUnauthorized
	}

	logger.WithField("code", code).Errorf("unknown error code")
	return http.StatusInternalServerError
}

type actionHandlerFunc func(r *http.Request) (interface{}, error)

func actionHandler(action actionHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		records, err := action(r)
		if err != nil {
			writeError(w, err)
			return
		}

		respondWith(w, records)
	}
}

//
// resourceRoute mounts resources in a convetional RESTful manner.
//
//   routes := resourceRoute{pathPrefix: "checks", router: ...}
//   routes.index(myIndexAction)    // given action is mounted at GET /checks
//   routes.show(myShowAction)      // given action is mounted at GET /checks/:id
//   routes.update(myUpdateAction)  // given action is mounted at {PUT,PATCH} /checks/:id
//   routes.create(myCreateAction)  // given action is mounted at POST /checks
//   routes.destroy(myCreateAction) // given action is mounted at DELETE /checks/:id
//   routes.path("{id}/publish", publishAction).Methods(http.MethodDelete) // when you need something customer
//
type resourceRoute struct {
	router     *mux.Router
	pathPrefix string
}

func (r *resourceRoute) index(fn actionHandlerFunc) *mux.Route {
	return r.path("", fn).Methods(http.MethodGet)
}

func (r *resourceRoute) show(fn actionHandlerFunc) *mux.Route {
	return r.path("{id}", fn).Methods(http.MethodGet)
}

func (r *resourceRoute) create(fn actionHandlerFunc) *mux.Route {
	return r.path("", fn).Methods(http.MethodPost)
}

func (r *resourceRoute) update(fn actionHandlerFunc) *mux.Route {
	return r.path("{id}", fn).Methods(http.MethodPut, http.MethodPatch)
}

func (r *resourceRoute) destroy(fn actionHandlerFunc) *mux.Route {
	return r.path("{id}", fn).Methods(http.MethodDelete)
}

func (r *resourceRoute) path(path string, fn actionHandlerFunc) *mux.Route {
	return handleAction(r.router, fmt.Sprintf("%s/%s", r.pathPrefix, path), fn)
}

func handleAction(router *mux.Router, path string, fn actionHandlerFunc) *mux.Route {
	return router.HandleFunc(path, actionHandler(fn))
}
