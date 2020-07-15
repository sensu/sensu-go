package pipeline

import (
	"context"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/api"
	"github.com/sensu/sensu-go/backend/authorization/rbac"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/types/dynamic"
	"github.com/stretchr/testify/mock"
)

func TestJavascriptStoreAccess(t *testing.T) {
	st := new(mockstore.MockStore)
	pipelineRoleBinding := &corev2.RoleBinding{
		Subjects: []corev2.Subject{
			{
				Type: corev2.GroupType,
				Name: "system:pipeline",
			},
		},
		RoleRef: corev2.RoleRef{
			Type: "Role",
			Name: "system:pipeline",
		},
		ObjectMeta: corev2.ObjectMeta{
			Namespace: "default",
			Name:      "system:pipeline",
		},
	}
	pipelineRole := &corev2.Role{
		Rules: []corev2.Rule{
			{
				Verbs: []string{
					"get", "list",
				},
				Resources: []string{
					"events",
				},
			},
		},
		ObjectMeta: corev2.ObjectMeta{
			Namespace: "default",
			Name:      "system:pipeline",
		},
	}
	event := corev2.FixtureEvent("entity", "check")

	// store mock supports rbac authorizer
	st.On("ListClusterRoleBindings", mock.Anything, mock.Anything).Return(([]*corev2.ClusterRoleBinding)(nil), nil)
	st.On("ListRoleBindings", mock.Anything, mock.Anything).Return([]*corev2.RoleBinding{pipelineRoleBinding}, nil)
	st.On("GetRole", mock.Anything, "system:pipeline").Return(pipelineRole, nil)
	st.On("GetClusterRole", mock.Anything).Return(nil, nil)

	// store mock supports event store
	st.On("GetEventByEntityCheck", mock.Anything, "entity", "check").Return(event, nil)
	st.On("GetEventByEntityCheck", mock.Anything, mock.Anything, mock.Anything).Return((*corev2.Event)(nil), nil)
	st.On("GetEvents", mock.Anything, mock.Anything).Return([]*corev2.Event{event}, nil)

	auth := &rbac.Authorizer{
		Store: st,
	}
	ctx := store.NamespaceContext(context.Background(), "default")
	client := api.NewEventClient(st, auth, nil)
	synthEvent := dynamic.Synthesize(event)
	funcs := map[string]interface{}{
		"FetchEvent": client.FetchEvent,
		"ListEvents": client.ListEvents,
	}
	env := FilterExecutionEnvironment{
		Event: synthEvent,
		Funcs: funcs,
	}
	tests := []struct {
		Name   string
		Expr   string
		Match  bool
		ExpErr string
	}{
		{
			Name: "lookup event, match on status 0",
			Expr: `(function () {
				var e = sensu.FetchEvent("entity", "check");
				return e.check.status == 0;
			})()`,
			Match: true,
		},
		{
			Name: "lookup event, match on status 1",
			Expr: `(function () {
				var e = sensu.FetchEvent("entity", "check");
				return e.check.status == 1;
			})()`,
			Match: false,
		},
		{
			Name: "error - nil event, undefined lookup",
			Expr: `(function () {
				var e = sensu.FetchEvent("batman", "robin");
				return e.check.status == 0;
			})()`,
			ExpErr: "TypeError: Cannot access member 'check' of undefined",
		},
		{
			Name: "list events",
			Expr: `(function () {
				var events = sensu.ListEvents();
				return events[0].check.status == 0;
			})()`,
			Match: true,
		},
		{
			Name: "no access to delete",
			Expr: `(function () {
				sensu.DeleteEvent("entity", "check");
				return true;
			})()`,
			ExpErr: "TypeError: 'DeleteEvent' is not a function",
		},
		{
			Name: "no access to update",
			Expr: `(function () {
				sensu.UpdateEvent({});
				return true;
			})()`,
			ExpErr: "TypeError: 'UpdateEvent' is not a function",
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			match, err := env.Eval(ctx, test.Expr)
			if err != nil && test.ExpErr == "" {
				t.Fatal(err)
			}
			if err != nil && test.ExpErr != "" {
				if got, want := err.Error(), test.ExpErr; got != want {
					t.Errorf("bad error: got %q, want %q", got, want)
				}
			}
			if test.ExpErr != "" && err == nil {
				t.Fatal("expected non-nil error")
			}
			if got, want := match, test.Match; got != want {
				t.Errorf("bad eval: got %v, want %v", got, want)
			}
		})
	}
	st.AssertCalled(t, "GetEventByEntityCheck", mock.Anything, "entity", "check")
	st.AssertCalled(t, "GetEventByEntityCheck", mock.Anything, "batman", "robin")
	st.AssertCalled(t, "GetEvents", mock.Anything, mock.Anything)
}
