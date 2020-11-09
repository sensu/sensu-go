package agentd

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/etcd"
	"github.com/sensu/sensu-go/backend/store"
	etcdstore "github.com/sensu/sensu-go/backend/store/etcd"
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
		storeErr     error
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
			storeErr:     fmt.Errorf("user not found"),
			expectedCode: http.StatusUnauthorized,
		},
	}

	for _, tc := range tests {
		stor := &mockstore.MockStore{}
		user := corev2.FixtureUser(tc.username)
		user.Groups = append(user.Groups, tc.group)
		stor.On("GetUser", mock.Anything, tc.username).Return(user, tc.storeErr)
		stor.On("AuthenticateUser", mock.Anything, tc.username, "password").Return(user, tc.storeErr)
		stor.On("ListClusterRoleBindings", mock.Anything, &store.SelectionPredicate{}).
			Return([]*corev2.ClusterRoleBinding{{
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
			}}, nil)
		stor.On("ListRoleBindings", mock.Anything, &store.SelectionPredicate{}).
			Return([]*corev2.RoleBinding{{
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
			}}, nil)
		stor.On("GetClusterRole", mock.Anything, "admin", mock.Anything).
			Return(&corev2.ClusterRole{Rules: []corev2.Rule{
				{
					Verbs:     []string{"create"},
					Resources: []string{"events"},
				},
			}}, nil)
		agentd := &Agentd{store: stor}
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
		watchEvent store.WatchEventEntityConfig
	}{
		{
			name: "bus error",
			watchEvent: store.WatchEventEntityConfig{
				Action: store.WatchCreate,
				Entity: corev3.FixtureEntityConfig("foo"),
			},
			busFunc: func(bus *mockbus.MockBus) {
				bus.On("Publish", mock.Anything, mock.Anything).Once().Return(errors.New("error"))
			},
		},
		{
			name: "watch events are successfully published to the bus",
			watchEvent: store.WatchEventEntityConfig{
				Action: store.WatchCreate,
				Entity: corev3.FixtureEntityConfig("foo"),
			},
			busFunc: func(bus *mockbus.MockBus) {
				bus.On("Publish", mock.Anything, mock.Anything).Once().Return(nil)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			watcher := make(chan store.WatchEventEntityConfig)

			// Mock the bus
			bus := &mockbus.MockBus{}
			if tt.busFunc != nil {
				tt.busFunc(bus)
			}

			e, cleanup := etcd.NewTestEtcd(t)
			defer cleanup()
			client := e.NewEmbeddedClient()
			defer client.Close()

			agent, err := New(Config{
				Bus:     bus,
				Watcher: watcher,
				Client:  client,
			})
			assert.NoError(t, err)

			go agent.runWatcher()

			watcher <- tt.watchEvent
		})
	}
}

func TestHealthHandler(t *testing.T) {
	e, cleanup := etcd.NewTestEtcd(t)
	defer cleanup()
	client := e.NewEmbeddedClient()
	defer client.Close()

	stor := etcdstore.NewStore(client, "test")
	agent, err := New(Config{
		Store:  stor,
		Client: client,
	})
	assert.NoError(t, err)

	srv := httptest.NewServer(agent.httpServer.Handler)
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/health", srv.URL), bytes.NewBuffer([]byte{}))
	res, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, res.StatusCode)
}
