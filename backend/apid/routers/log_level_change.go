package routers

import (
	"context"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/util/logging"
)

// LogLevelChangeController represents the controller needs of the LogLevelChangeRouter.
type LogLevelChangeController interface {
	Create(ctx context.Context, requests []logging.LogLevelRequest) ([]logging.LogHistory, error)
	List(ctx context.Context) ([]logging.LogLevelRequest, error)
	ModuleLogLevel(ctx context.Context, module string) (*logging.LogLevelRequest, error)
	SetGlobalModuleLogLevel(ctx context.Context, logLevel string) ([]*logging.LogHistory, error)
}

// LogLevelChangeRouter handles requests for /loglevel
type LogLevelChangeRouter struct {
	controller LogLevelChangeController
}

// NewLogLevelChangeRouter instantiates new router for controlling user resources
func NewLogLevelChangeRouter() *LogLevelChangeRouter {
	return &LogLevelChangeRouter{
		controller: actions.NewLogLevelChangeController(),
	}
}

// Mount the LogLevelChangeRouter to a parent Router
func (r *LogLevelChangeRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/{resource:loglevel}",
	}

	routes.Post(r.create)
	routes.Path("{subresource:modules}", r.list).Methods(http.MethodGet)
	routes.Path("{subresource:modules}/{loglevel}", r.SetGlobalModuleLogLevel).Methods(http.MethodPost)
	routes.Path("{subresource:modules}/{module}", r.moduleLogLevel).Methods(http.MethodGet)
}

// create function will change log level of a module
func (r *LogLevelChangeRouter) create(req *http.Request) (interface{}, error) {
	var logReqs []logging.LogLevelRequest
	if err := UnmarshalBody(req, &logReqs); err != nil {
		return nil, actions.NewError(actions.InvalidArgument, err)
	}

	changeLog, err := r.controller.Create(req.Context(), logReqs)
	return changeLog, err
}

func (r *LogLevelChangeRouter) SetGlobalModuleLogLevel(req *http.Request) (interface{}, error) {
	vars := mux.Vars(req)
	logLevel, err := url.PathUnescape(vars["loglevel"])
	if err != nil {
		return nil, err
	}
	moduleLogLevel, err := r.controller.SetGlobalModuleLogLevel(req.Context(), logLevel)
	return moduleLogLevel, err
}

// moduleLogLevel function gives an module log level
func (r *LogLevelChangeRouter) moduleLogLevel(req *http.Request) (interface{}, error) {
	vars := mux.Vars(req)
	moduleName, err := url.PathUnescape(vars["module"])
	if err != nil {
		return nil, err
	}
	moduleLogLevel, err := r.controller.ModuleLogLevel(req.Context(), moduleName)
	return moduleLogLevel, err
}

// list function gives list of all modules registered with corresponding log level
func (r *LogLevelChangeRouter) list(req *http.Request) (interface{}, error) {
	moduleList, err := r.controller.List(req.Context())
	return moduleList, err
}
