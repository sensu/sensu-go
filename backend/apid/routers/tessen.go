package routers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	"github.com/sensu/sensu-go/backend/apid/request"
)

// TessenController represents the controller needs of the TessenRouter.
type TessenController interface {
	CreateOrUpdate(context.Context, *corev2.TessenConfig) error
	Get(context.Context) (*corev2.TessenConfig, error)
}

// TessenRouter handles requests for /tessen.
type TessenRouter struct {
	controller TessenController
}

// NewTessenRouter instantiates a new router for tessen.
func NewTessenRouter(ctrl TessenController) *TessenRouter {
	return &TessenRouter{
		controller: ctrl,
	}
}

// Mount the TessenRouter on the given parent Router
func (r *TessenRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/" + corev2.TessenResource,
	}

	routes.Path("", r.get).Methods(http.MethodGet)
	routes.Path("", r.createOrUpdate).Methods(http.MethodPut)
}

func (r *TessenRouter) createOrUpdate(req *http.Request) (handlers.HandlerResponse, error) {
	var response handlers.HandlerResponse
	obj, err := request.Resource[*corev2.TessenConfig](req)
	if err != nil {
		return response, err
	}

	err = r.controller.CreateOrUpdate(req.Context(), obj)
	response.Resource = obj
	return response, err
}

func (r *TessenRouter) get(req *http.Request) (handlers.HandlerResponse, error) {
	return responseWrap(r.controller.Get(req.Context()))
}

// TessenMetricController represents the controller needs of the TessenMetricRouter.
type TessenMetricController interface {
	Publish(context.Context, []corev2.MetricPoint) error
}

// TessenMetricRouter handles requests for /tessen/metrics.
type TessenMetricRouter struct {
	controller TessenMetricController
}

// NewTessenMetricRouter instantiates a new router for tessen metrics.
func NewTessenMetricRouter(ctrl TessenMetricController) *TessenMetricRouter {
	return &TessenMetricRouter{
		controller: ctrl,
	}
}

// Mount the TessenMetricRouter on the given parent Router
func (r *TessenMetricRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/api/{group:core}/{version:v2}/tessen/metrics",
	}

	routes.Path("", r.publish).Methods(http.MethodPost)
}

func (r *TessenMetricRouter) publish(req *http.Request) (handlers.HandlerResponse, error) {
	var obj []corev2.MetricPoint
	var response handlers.HandlerResponse
	if err := json.NewDecoder(req.Body).Decode(&obj); err != nil {
		return response, err
	}

	err := r.controller.Publish(req.Context(), obj)
	return response, err
}
