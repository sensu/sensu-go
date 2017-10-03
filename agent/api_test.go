package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
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
			config := NewConfig()
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
