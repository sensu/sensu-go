package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/sensu/sensu-go/testing/mocktransport"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestAddEvent(t *testing.T) {
	testCases := []struct {
		desc             string
		event            interface{}
		expectedResponse int
	}{
		{
			"with an empty event",
			nil,
			http.StatusBadRequest,
		},
		{
			"with a bad event",
			interface{}("foo"),
			http.StatusBadRequest,
		},
		{
			"with an event without an entity",
			types.Event{
				Check: types.FixtureCheck("check_foo"),
			},
			http.StatusCreated,
		},
		{
			"with a proper event",
			types.FixtureEvent("foo", "check_foo"),
			http.StatusCreated,
		},
	}

	for _, tc := range testCases {
		testName := fmt.Sprintf("add event %s", tc.desc)
		t.Run(testName, func(t *testing.T) {
			config := FixtureConfig()
			agent := NewAgent(config)

			encoded, _ := json.Marshal(tc.event)
			r, err := http.NewRequest("POST", "/events", bytes.NewBuffer(encoded))
			assert.NoError(t, err)

			router := mux.NewRouter()
			registerRoutes(agent, router)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)

			assert.Equal(t, tc.expectedResponse, w.Code)
		})
	}
}

func TestHealthz(t *testing.T) {
	testCases := []struct {
		desc             string
		expectedResponse int
		closeConn        bool
	}{
		{
			"healthz returns success",
			http.StatusOK,
			false,
		},
		{
			"healthz returns failure",
			http.StatusServiceUnavailable,
			true,
		},
	}

	for _, tc := range testCases {
		testName := fmt.Sprintf("test agent %s", tc.desc)
		transport := &mocktransport.MockTransport{}

		t.Run(testName, func(t *testing.T) {
			// need to figure out how to pass the mock transport into the agent
			config := FixtureConfig()
			agent := NewAgent(config)
			agent.conn = transport
			transport.On("Closed").Return(tc.closeConn)

			r, err := http.NewRequest("GET", "/healthz", nil)
			assert.NoError(t, err)

			router := mux.NewRouter()
			registerRoutes(agent, router)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, r)

			assert.Equal(t, tc.expectedResponse, w.Code)
		})
	}
}
