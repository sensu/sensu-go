package agentd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/apid/routers"
	"github.com/sensu/sensu-go/backend/etcd"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/etcdstore"
	"github.com/sensu/sensu-go/backend/store/v2/wrap"
	"github.com/sensu/sensu-go/testing/mockbus"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/transport"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestAgentdMiddlewares(t *testing.T) {
	assert := assert.New(t)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	tests := []struct {
		description  string
		namespace    string
		agentName    string
		username     string
		group        string
		authenErr    error
		expectedCode int
	}{
		{
			description:  "Authorized request",
			namespace:    "test-rbac",
			username:     "authorized-user",
			group:        "group-test-rbac",
			expectedCode: http.StatusOK,
		}, {
			description:  "Unauthorized request",
			namespace:    "super-secret",
			username:     "unauthorized-user",
			expectedCode: http.StatusForbidden,
		}, {
			description:  "Invalid user",
			namespace:    "test-rbac",
			username:     "nonexistent-user",
			authenErr:    fmt.Errorf("user not found"),
			expectedCode: http.StatusUnauthorized,
		},
	}

	resourceRequestMatcher := func(storeName, namespace, name string) func(interface{}) bool {
		return func(i interface{}) bool {
			rr, ok := i.(storev2.ResourceRequest)
			if !ok {
				return false
			}

			if rr.StoreName == storeName && rr.Name == name {
				switch rr.StoreName {
				case "cluster_roles", "cluster_role_bindings":
					// These resources are not namespaced so we ignore the
					// Namespace field altogether.
					return true

				default:
					return rr.Namespace == namespace
				}
			}

			return false
		}
	}

	for _, tc := range tests {
		authenticator := &mockAuthenticator{}
		claims := corev2.FixtureClaims(tc.username, []string{"default", tc.group})
		authenticator.On("Authenticate", mock.Anything, tc.username, "password").Return(claims, tc.authenErr)
		stor := &mockstore.V2MockStore{}
		stor.On("List", mock.Anything, mock.MatchedBy(resourceRequestMatcher("cluster_role_bindings", tc.namespace, "")), mock.Anything).
			Return(mockstore.WrapList[*corev2.ClusterRoleBinding](
				[]*corev2.ClusterRoleBinding{{
					RoleRef: corev2.RoleRef{
						Type: "ClusterRole",
						Name: "cluster-admin",
					},
					Subjects: []corev2.Subject{
						{Type: corev2.GroupType, Name: "cluster-admins"},
					},
					ObjectMeta: corev2.ObjectMeta{
						Name: "cluster-admin",
					},
				}}), nil)

		stor.On("List", mock.Anything, mock.MatchedBy(resourceRequestMatcher("role_bindings", tc.namespace, "")), mock.Anything).
			Return(mockstore.WrapList[*corev2.RoleBinding](
				[]*corev2.RoleBinding{{
					RoleRef: corev2.RoleRef{
						Type: "ClusterRole",
						Name: "admin",
					},
					Subjects: []corev2.Subject{
						{Type: corev2.GroupType, Name: "group-test-rbac"},
					},
					ObjectMeta: corev2.ObjectMeta{
						Name:      "role-test-rbac-admin",
						Namespace: "test-rbac",
					},
				}}), nil)
		stor.On("Get", mock.Anything, mock.MatchedBy(resourceRequestMatcher("cluster_roles", tc.namespace, "admin")), mock.Anything).
			Return(wrapResource(t,
				&corev2.ClusterRole{
					ObjectMeta: corev2.NewObjectMeta("group-test-rbac", ""),
					Rules: []corev2.Rule{
						{
							Verbs:     []string{"create"},
							Resources: []string{"events"},
						},
					}}), nil)

		agentd := &Agentd{store: stor, authenticator: authenticator}
		server := httptest.NewServer(agentd.AuthenticationMiddleware(agentd.AuthorizationMiddleware(testHandler)))
		defer server.Close()
		req, _ := http.NewRequest(http.MethodPost, server.URL, bytes.NewBuffer([]byte{}))
		req.SetBasicAuth(tc.username, "password")
		req.Header.Set(transport.HeaderKeyNamespace, tc.namespace)
		req.Header.Set(transport.HeaderKeyAgentName, tc.agentName)
		req.Header.Set(transport.HeaderKeyUser, tc.username)
		res, err := http.DefaultClient.Do(req)
		assert.NoError(err)
		assert.Equal(tc.expectedCode, res.StatusCode, tc.description)
	}
}

func TestRunWatcher(t *testing.T) {
	type busFunc func(*mockbus.MockBus)

	tests := []struct {
		name       string
		busFunc    busFunc
		watchEvent []storev2.WatchEvent
	}{
		{
			name: "bus error",
			watchEvent: []storev2.WatchEvent{
				{
					Type:  storev2.WatchCreate,
					Value: wrapResource(t, corev3.FixtureEntityConfig("foo")),
				},
			},
			busFunc: func(bus *mockbus.MockBus) {
				bus.On("Publish", mock.Anything, mock.Anything).Once().Return(errors.New("error"))
			},
		},
		{
			name: "watch events are successfully published to the bus",
			watchEvent: []storev2.WatchEvent{
				{
					Type:  storev2.WatchCreate,
					Value: wrapResource(t, corev3.FixtureEntityConfig("foo")),
				},
			},
			busFunc: func(bus *mockbus.MockBus) {
				bus.On("Publish", mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			watcher := make(chan []storev2.WatchEvent)

			// Mock the bus
			bus := &mockbus.MockBus{}
			if tt.busFunc != nil {
				tt.busFunc(bus)
			}

			e, cleanup := etcd.NewTestEtcd(t)
			defer cleanup()
			client := e.NewEmbeddedClient()
			defer func() { _ = client.Close() }()
			stor := etcdstore.NewStore(client)

			agent, err := New(Config{
				Bus:          bus,
				Watcher:      watcher,
				Store:        stor,
				HealthRouter: routers.NewHealthRouter(nil),
			})
			assert.NoError(t, err)

			go agent.runWatcher()

			watcher <- tt.watchEvent
		})
	}
}

type mockAuthenticator struct {
	mock.Mock
}

func (m *mockAuthenticator) Authenticate(ctx context.Context, username, password string) (*corev2.Claims, error) {
	args := m.Called(ctx, username, password)
	claims, _ := args.Get(0).(*corev2.Claims)
	return claims, args.Error(1)
}

func wrapResource(t *testing.T, r corev3.Resource, opts ...wrap.Option) storev2.Wrapper {
	w, err := storev2.WrapResource(r, opts...)
	if err != nil {
		t.Errorf("error wrapping resource: %v", err)
	}
	return w
}
