package routers

import (
	"encoding/json"
	"io"
	"net/http"
	"path"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/actions"
)

type errorBody struct {
	Message string `json:"message"`
	Code    uint32 `json:"code"`
}

// FilterFunc ...
type FilterFunc func(w http.ResponseWriter, r *http.Request, resources interface{}) error

// Filter ...
var Filter FilterFunc

func init() {
	Filter = func(w http.ResponseWriter, r *http.Request, resources interface{}) error {
		return nil
	}
}

// respondWith given writer and resource, marshal to JSON and write response.
func respondWith(w http.ResponseWriter, resources interface{}) {
	// Set content-type to JSON
	w.Header().Set("Content-Type", "application/json")

	// If no resource(s) are present return a 204 response code
	if resources == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Marshal
	bytes, err := json.Marshal(resources)
	if err != nil {
		writeError(w, err)
		return
	}

	// Write response
	if _, err := w.Write(bytes); err != nil {
		logger.WithError(err).Error("failed to write response")
		writeError(w, err)
	}
}

// writeError writes error response in JSON format.
func writeError(w http.ResponseWriter, err error) {
	const fallback = `{"message": "failed to marshal error message"}`

	errBody := errorBody{}
	st := http.StatusInternalServerError

	// Wrap message in standard errorBody
	actionErr, ok := err.(actions.Error)
	if ok {
		errBody.Message = actionErr.Message
		errBody.Code = uint32(actionErr.Code)
		st = HTTPStatusFromCode(actionErr.Code)
	} else {
		errBody.Message = err.Error()
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

	// Write error message in JSON encoded message
	w.WriteHeader(st)
	_, _ = w.Write(errJSON)
}

// HTTPStatusFromCode returns http status code for given user action err code
func HTTPStatusFromCode(code actions.ErrCode) int {
	switch code {
	case actions.InternalErr:
		return http.StatusInternalServerError
	case actions.InvalidArgument:
		return http.StatusBadRequest
	case actions.NotFound:
		return http.StatusNotFound
	case actions.AlreadyExistsErr:
		return http.StatusConflict
	}

	logger.WithField("code", code).Error("unknown error code")
	return http.StatusInternalServerError
}

//
// actionHandler takes a action handler closure and returns a new handler that
// exexutes the closure and writes the response.
//
// Ex.
//
//   handler := actionHandler(func(r *http.Request) (interface{}, error) {
//     msg := r.Vars("message")
//     if msg == "i-am-a-jerk" {
//       return nil, errors.New("fatal err")
//     }
//     return strings.Split(msg, "-"), nil
//   })
//   router.handleFunc("/echo/{message}", handler).Methods(http.MethodGet)
//
//    GET /echo/hey         --> 200 OK ["hey"]
//    GET /echo/hey-there   --> 200 OK ["howdy", "there"]
//    GET /echo/i-am-a-jerk --> 500    {code: 500, message: "fatal err"}
//
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

func listHandler(fn listHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		records, err := fn(w, r)
		if err != nil {
			writeError(w, err)
			return
		}

		_ = Filter(w, r, records)

		respondWith(w, records)
	}
}

type actionHandlerFunc func(r *http.Request) (interface{}, error)
type listHandlerFunc func(w http.ResponseWriter, req *http.Request) (interface{}, error)

//
// ResourceRoute mounts resources in a convetional RESTful manner.
//
//   routes := ResourceRoute{PathPrefix: "checks", Router: ...}
//   routes.Get(myShowAction)     // given action is mounted at GET /checks/:id
//   routes.List(myIndexAction)   // given action is mounted at GET /checks
//   routes.Put(myCreateAction)   // given action is mounted at PUT /checks/:id
//   routes.Patch(myUpdateAction) // given action is mounted at PATCH /checks/:id
//   routes.Post(myCreateAction)  // given action is mounted at POST /checks
//   routes.Del(myCreateAction)   // given action is mounted at DELETE /checks/:id
//   routes.Path("{id}/publish", publishAction).Methods(http.MethodDelete) // when you need something customer
//
type ResourceRoute struct {
	Router     *mux.Router
	PathPrefix string
}

// Get reads
func (r *ResourceRoute) Get(fn actionHandlerFunc) *mux.Route {
	return r.Path("{id}", fn).Methods(http.MethodGet)
}

// List resources
func (r *ResourceRoute) List(fn listHandlerFunc) *mux.Route {
	return r.Router.HandleFunc(r.PathPrefix, listHandler(fn)).Methods(http.MethodGet)
}

// ListAllNamespaces return all resources across all namespaces
func (r *ResourceRoute) ListAllNamespaces(fn listHandlerFunc, path string) *mux.Route {
	return r.Router.HandleFunc(path, listHandler(fn)).Methods(http.MethodGet)
}

// Post creates
func (r *ResourceRoute) Post(fn actionHandlerFunc) *mux.Route {
	return r.Path("", fn).Methods(http.MethodPost)
}

// TODO: uncomment this and use it once controller update fits
// http patch semantics.
// Patch updates/modifies
//func (r *ResourceRoute) Patch(fn actionHandlerFunc) *mux.Route {
//	return r.path("{id}", fn).Methods(http.MethodPatch)
//}

// Put updates/replaces
func (r *ResourceRoute) Put(fn actionHandlerFunc) *mux.Route {
	return r.Path("{id}", fn).Methods(http.MethodPut)
}

// Del deletes
func (r *ResourceRoute) Del(fn actionHandlerFunc) *mux.Route {
	return r.Path("{id}", fn).Methods(http.MethodDelete)
}

// Path adds custom path
func (r *ResourceRoute) Path(p string, fn actionHandlerFunc) *mux.Route {
	fullPath := path.Join(r.PathPrefix, p)
	return handleAction(r.Router, fullPath, fn)
}

func handleAction(router *mux.Router, path string, fn actionHandlerFunc) *mux.Route {
	return router.HandleFunc(path, actionHandler(fn))
}

// UnmarshalBody decodes the request body
func UnmarshalBody(req *http.Request, record interface{}) error {
	err := json.NewDecoder(req.Body).Decode(&record)
	if err != nil {
		logger.WithError(err).Error("unable to read request body")
		return err
	}
	// TODO: Support other types of requests other than JSON?

	return nil
}
