package routers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/middlewares"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockGenericController struct {
	mock.Mock
}

func (m *mockGenericController) List(ctx context.Context, pred *store.SelectionPredicate) ([]corev2.Resource, error) {
	args := m.Called(ctx, pred)
	return args.Get(0).([]corev2.Resource), args.Error(1)
}

func TestList(t *testing.T) {
	tests := []struct {
		name                   string
		path                   string
		results                []corev2.Resource
		controllerErr          error
		continueToken          string
		expectedContinueHeader string
		expectedLen            int
		expectedPred           *store.SelectionPredicate
		expectedStatus         int
	}{
		{
			name:           "list without pagination",
			path:           "/foo",
			results:        []corev2.Resource{corev2.FixtureCheck("check-cpu")},
			expectedLen:    1,
			expectedPred:   &store.SelectionPredicate{},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "list with subcollection but without pagination",
			path:           "/foo/bar",
			results:        []corev2.Resource{corev2.FixtureCheck("check-cpu")},
			expectedLen:    1,
			expectedPred:   &store.SelectionPredicate{Subcollection: "bar"},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "controller error",
			path:           "/foo",
			controllerErr:  errors.New("error"),
			expectedPred:   &store.SelectionPredicate{},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:                   "continue token",
			path:                   "/foo",
			results:                []corev2.Resource{corev2.FixtureCheck("check-cpu")},
			continueToken:          "bar",
			expectedLen:            1,
			expectedPred:           &store.SelectionPredicate{},
			expectedStatus:         http.StatusOK,
			expectedContinueHeader: "YmFy",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := &mockGenericController{}
			controller.On("List", mock.Anything, mock.AnythingOfType("*store.SelectionPredicate")).
				Return(tt.results, tt.controllerErr).
				Run(func(args mock.Arguments) {
					pred := args[1].(*store.SelectionPredicate)
					assert.Equal(t, tt.expectedPred, pred)

					if tt.continueToken != "" {
						pred.Continue = tt.continueToken
					}
				})

			r, err := http.NewRequest("GET", tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()

			router := mux.NewRouter()
			router.PathPrefix("/foo/{subcollection}").HandlerFunc(List(controller.List,
				func(r corev2.Resource) map[string]string { return map[string]string{} },
			))
			router.PathPrefix("/foo").HandlerFunc(List(controller.List,
				func(r corev2.Resource) map[string]string { return map[string]string{} },
			))
			middleware := middlewares.Pagination{}
			router.Use(middleware.Then)
			router.ServeHTTP(w, r)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if w.Code < 400 {
				payload := []interface{}{}
				if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
					t.Fatal(err)
				}
				assert.Len(t, payload, tt.expectedLen)
			}
			assert.Equal(t, tt.expectedContinueHeader, w.Header().Get(corev2.PaginationContinueHeader))
		})
	}
}
