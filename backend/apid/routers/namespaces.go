package routers

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	"github.com/sensu/sensu-go/backend/authorization"
	"github.com/sensu/sensu-go/backend/authorization/rbac"
	"github.com/sensu/sensu-go/backend/store"
)

type ListStore interface {
	ListResources(ctx context.Context, kind string, resources interface{}, pred *store.SelectionPredicate) error
}

type RuleVisitor interface {
	VisitRulesFor(ctx context.Context, attrs *authorization.Attributes, fn rbac.RuleVisitFunc)
}

// NamespacesRouter handles requests for /namespaces
type NamespacesRouter struct {
	handlers handlers.Handlers
	store    ListStore
	auth     RuleVisitor
}

// NewNamespacesRouter instantiates new router for controlling check resources
func NewNamespacesRouter(store store.ResourceStore, auth RuleVisitor) *NamespacesRouter {
	return &NamespacesRouter{
		store: store,
		auth:  auth,
		handlers: handlers.Handlers{
			Resource: &corev2.Namespace{},
			Store:    store,
		},
	}
}

// Mount the NamespacesRouter to a parent Router
func (r *NamespacesRouter) Mount(parent *mux.Router) {
	routes := ResourceRoute{
		Router:     parent,
		PathPrefix: "/{resource:namespaces}",
	}

	routes.Del(r.handlers.DeleteResource)
	routes.Get(r.get)
	routes.List(r.handlers.ListResources, corev2.NamespaceFields)
	routes.Post(r.handlers.CreateResource)
	routes.Put(r.handlers.CreateOrUpdateResource)
}

func (r *NamespacesRouter) get(req *http.Request) (interface{}, error) {
	attrs := authorization.GetAttributes(req.Context())
	if attrs == nil {
		return nil, actions.NewErrorf(actions.InvalidArgument)
	}

	resources := []*corev2.Namespace{}
	err := r.store.ListResources(req.Context(), corev2.NamespacesResource, &resources, &store.SelectionPredicate{})
	if err != nil {
		return nil, actions.NewError(actions.InternalErr, err)
	}

	namespaceMap := make(map[string]*corev2.Namespace, len(resources))
	for _, namespace := range resources {
		namespaceMap[namespace.Name] = namespace
	}

	namespaces := make([]*corev2.Namespace, 0, len(resources))

	var funcErr error

	// Iterate over the rbac rules to discover namespaces that this
	// user has read access to.
	r.auth.VisitRulesFor(req.Context(), attrs, func(binding rbac.RoleBinding, rule corev2.Rule, err error) (terminate bool) {
		if err != nil {
			funcErr = err
			return true
		}
		if len(namespaceMap) == 0 {
			return true
		}
		if !rule.VerbMatches("get") {
			return false
		}
		if !rule.ResourceMatches(corev2.NamespacesResource) {
			return false
		}
		if len(rule.ResourceNames) == 0 {
			// All resources of type "namespace" are allowed
			namespaces = resources
			return true
		}
		for name, namespace := range namespaceMap {
			if rule.ResourceNameMatches(name) {
				namespaces = append(namespaces, namespace)
				delete(namespaceMap, name)
			}
		}
		return false
	})

	if funcErr != nil {
		return nil, actions.NewErrorf(actions.InternalErr, funcErr)
	}

	return namespaces, nil
}
