package routers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/mock"
)

type storeFunc func(*mockstore.MockStore)

type routerTestCase struct {
	name           string
	method         string
	path           string
	body           []byte
	storeFunc      storeFunc
	wantStatusCode int
}

func run(t *testing.T, tt routerTestCase, router *mux.Router, store *mockstore.MockStore) bool {
	t.Helper()
	return t.Run(tt.name, func(t *testing.T) {
		// Only start the HTTP server here to prevent data races in tests
		t.Helper()
		server := httptest.NewServer(router)
		defer server.Close()

		if tt.storeFunc != nil {
			tt.storeFunc(store)
		}

		// Prepare the HTTP request
		client := new(http.Client)
		req, err := http.NewRequest(tt.method, server.URL+tt.path, bytes.NewReader(tt.body))
		if err != nil {
			t.Fatal(err)
		}

		// Perform the HTTP request
		res, err := client.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()

		// Inspect the response code
		if res.StatusCode != tt.wantStatusCode {
			t.Errorf("StatusCode = %v, wantStatusCode %v", res.StatusCode, tt.wantStatusCode)
			body, _ := ioutil.ReadAll(res.Body)
			t.Errorf("error message: %q", string(body))
			return
		}
	})
}

// Get
var getTestCases = func(resource corev2.Resource) []routerTestCase {
	return []routerTestCase{
		getResourceNotFoundTestCase(resource),
		getResourceStoreErrTestCase(resource),
		getResourceSuccessTestCase(resource),
	}
}

var getResourceNotFoundTestCase = func(resource corev2.Resource) routerTestCase {
	name := resource.GetObjectMeta().Name
	typ := reflect.TypeOf(resource).String()

	return routerTestCase{
		name:   "it returns 404 if a resource is not found",
		method: http.MethodGet,
		path:   resource.URIPath(),
		storeFunc: func(s *mockstore.MockStore) {
			s.On("GetResource", mock.Anything, name, mock.AnythingOfType(typ)).
				Return(&store.ErrNotFound{}).
				Once()
		},
		wantStatusCode: http.StatusNotFound,
	}
}

var getResourceStoreErrTestCase = func(resource corev2.Resource) routerTestCase {
	name := resource.GetObjectMeta().Name
	typ := reflect.TypeOf(resource).String()

	return routerTestCase{
		name:   "it returns 500 if the store encounters an error while retrieving a resource",
		method: http.MethodGet,
		path:   resource.URIPath(),
		storeFunc: func(s *mockstore.MockStore) {
			s.On("GetResource", mock.Anything, name, mock.AnythingOfType(typ)).
				Return(&store.ErrInternal{}).
				Once()
		},
		wantStatusCode: http.StatusInternalServerError,
	}
}

var getResourceSuccessTestCase = func(resource corev2.Resource) routerTestCase {
	name := resource.GetObjectMeta().Name
	typ := reflect.TypeOf(resource).String()

	return routerTestCase{
		name:   "it retrieves a check config",
		method: http.MethodGet,
		path:   resource.URIPath(),
		storeFunc: func(s *mockstore.MockStore) {
			s.On("GetResource", mock.Anything, name, mock.AnythingOfType(typ)).
				Return(nil).
				Once()
		},
		wantStatusCode: http.StatusOK,
	}
}

// List
var listTestCases = func(resource corev2.Resource) []routerTestCase {
	return []routerTestCase{
		listResourcesStoreErrTestCase(resource),
		listResourcesSuccessTestCase(resource),
		listResourcesAcrossNamespacesTestCase(resource),
	}
}

var listResourcesStoreErrTestCase = func(resource corev2.Resource) routerTestCase {
	resource.SetNamespace("default")
	typ := reflect.TypeOf(resource).String()

	return routerTestCase{
		name:   "it returns 500 if the store encounters an error while listing resources",
		method: http.MethodGet,
		path:   resource.URIPath(),
		storeFunc: func(s *mockstore.MockStore) {
			s.On("ListResources", mock.Anything, resource.StorePrefix(), mock.AnythingOfType("*[]"+typ), mock.AnythingOfType("*store.SelectionPredicate")).
				Return(&store.ErrInternal{}).
				Once()
		},
		wantStatusCode: http.StatusInternalServerError,
	}
}

var listResourcesSuccessTestCase = func(resource corev2.Resource) routerTestCase {
	resource.SetNamespace("default")
	typ := reflect.TypeOf(resource).String()

	return routerTestCase{
		name:   "it lists resources",
		method: http.MethodGet,
		path:   resource.URIPath(),
		storeFunc: func(s *mockstore.MockStore) {
			s.On("ListResources", mock.Anything, resource.StorePrefix(), mock.AnythingOfType("*[]"+typ), mock.AnythingOfType("*store.SelectionPredicate")).
				Return(nil).
				Once()
		},
		wantStatusCode: http.StatusOK,
	}
}

var listResourcesAcrossNamespacesTestCase = func(resource corev2.Resource) routerTestCase {
	typ := reflect.TypeOf(resource).String()
	return routerTestCase{
		name:   "it lists resources across namespaces",
		method: http.MethodGet,
		path:   resource.URIPath(),
		storeFunc: func(s *mockstore.MockStore) {
			s.On("ListResources", mock.Anything, resource.StorePrefix(), mock.AnythingOfType("*[]"+typ), mock.AnythingOfType("*store.SelectionPredicate")).
				Return(nil).
				Once()
		},
		wantStatusCode: http.StatusOK,
	}
}

// Create
var createTestCases = func(resource corev2.Resource) []routerTestCase {
	return []routerTestCase{
		createResourceInvalidPayloadTestCase(resource),
		createResourceInvalidMetaTestCase(resource),
		createResourceAlreadyExistsTestCase(resource),
		createResourceInvalidTestCase(resource),
		createResourceStoreErrTestCase(resource),
		createResourceSuccessTestCase(resource),
	}
}

var createResourceInvalidPayloadTestCase = func(resource corev2.Resource) routerTestCase {
	resource.SetNamespace("default")

	return routerTestCase{
		name:           "it returns 400 if the request payload to create is invalid",
		method:         http.MethodPost,
		path:           resource.URIPath(),
		body:           []byte("foo"),
		wantStatusCode: http.StatusBadRequest,
	}
}

var createResourceInvalidMetaTestCase = func(resource corev2.Resource) routerTestCase {
	resource.SetNamespace("default")

	// Do not test this on global resources because it will fail
	if !strings.Contains(resource.URIPath(), "namespaces/default") {
		return routerTestCase{wantStatusCode: 404}
	}

	return routerTestCase{
		name:           "it returns 400 if the resource metadata to create does not match the request path",
		method:         http.MethodPost,
		path:           resource.URIPath(),
		body:           []byte(`{"metadata": {"namespace":"acme"}}`),
		wantStatusCode: http.StatusBadRequest,
	}
}

var createResourceAlreadyExistsTestCase = func(resource corev2.Resource) routerTestCase {
	resource.SetNamespace("default")
	typ := reflect.TypeOf(resource).String()

	return routerTestCase{
		name:   "it returns 409 if the resource to create already exists",
		method: http.MethodPost,
		path:   resource.URIPath(),
		body:   marshal(resource),
		storeFunc: func(s *mockstore.MockStore) {
			s.On("CreateResource", mock.Anything, mock.AnythingOfType(typ)).
				Return(&store.ErrAlreadyExists{}).
				Once()
		},
		wantStatusCode: http.StatusConflict,
	}
}

var createResourceInvalidTestCase = func(resource corev2.Resource) routerTestCase {
	resource.SetNamespace("default")
	typ := reflect.TypeOf(resource).String()

	return routerTestCase{
		name:   "it returns 400 if the resource to create is invalid",
		method: http.MethodPost,
		path:   resource.URIPath(),
		body:   marshal(resource),
		storeFunc: func(s *mockstore.MockStore) {
			s.On("CreateResource", mock.Anything, mock.AnythingOfType(typ)).
				Return(&store.ErrNotValid{Err: errors.New("error")}).
				Once()
		},
		wantStatusCode: http.StatusBadRequest,
	}
}

var createResourceStoreErrTestCase = func(resource corev2.Resource) routerTestCase {
	resource.SetNamespace("default")
	typ := reflect.TypeOf(resource).String()

	return routerTestCase{
		name:   "it returns 500 if the store returns an error while creating",
		method: http.MethodPost,
		path:   resource.URIPath(),
		body:   []byte(`{"metadata": {"namespace":"default","name":"foo"}}`),
		storeFunc: func(s *mockstore.MockStore) {
			s.On("CreateResource", mock.Anything, mock.AnythingOfType(typ)).
				Return(&store.ErrInternal{}).
				Once()
		},
		wantStatusCode: http.StatusInternalServerError,
	}
}

var createResourceSuccessTestCase = func(resource corev2.Resource) routerTestCase {
	resource.SetNamespace("default")
	typ := reflect.TypeOf(resource).String()

	return routerTestCase{
		name:   "it returns 201 if the resource was created",
		method: http.MethodPost,
		path:   resource.URIPath(),
		body:   marshal(resource),
		storeFunc: func(s *mockstore.MockStore) {
			s.On("CreateResource", mock.Anything, mock.AnythingOfType(typ)).
				Return(nil).
				Once()
		},
		wantStatusCode: http.StatusCreated,
	}
}

// Update
var updateTestCases = func(resource corev2.Resource) []routerTestCase {
	return []routerTestCase{
		updateResourceInvalidPayloadTestCase(resource),
		updateResourceInvalidMetaTestCase(resource),
		updateResourceInvalidTestCase(resource),
		updateResourceStoreErrTestCase(resource),
		updateResourceSuccessTestCase(resource),
	}
}

var updateResourceInvalidPayloadTestCase = func(resource corev2.Resource) routerTestCase {
	return routerTestCase{
		name:           "it returns 400 if the request payload to update is invalid",
		method:         http.MethodPut,
		path:           resource.URIPath(),
		body:           []byte("foo"),
		wantStatusCode: http.StatusBadRequest,
	}
}

var updateResourceInvalidMetaTestCase = func(resource corev2.Resource) routerTestCase {
	// fmt.Println(resource.URIPath())
	// body := marshal(resource)
	resource.SetNamespace("acme")
	// fmt.Println(string(body))

	return routerTestCase{
		name:           "it returns 400 if the resource metadata to update does not match the request path",
		method:         http.MethodPut,
		path:           resource.URIPath(),
		body:           []byte(`{"metadata":{"name":"bar"}}`),
		wantStatusCode: http.StatusBadRequest,
	}
}

var updateResourceInvalidTestCase = func(resource corev2.Resource) routerTestCase {
	typ := reflect.TypeOf(resource).String()

	return routerTestCase{
		name:   "it returns 400 if the resource to update is invalid",
		method: http.MethodPut,
		path:   resource.URIPath(),
		body:   marshal(resource),
		storeFunc: func(s *mockstore.MockStore) {
			s.On("CreateOrUpdateResource", mock.Anything, mock.AnythingOfType(typ)).
				Return(&store.ErrNotValid{Err: errors.New("error")}).
				Once()
		},
		wantStatusCode: http.StatusBadRequest,
	}
}

var updateResourceStoreErrTestCase = func(resource corev2.Resource) routerTestCase {
	typ := reflect.TypeOf(resource).String()

	return routerTestCase{
		name:   "it returns 500 if the store returns an error while updating",
		method: http.MethodPut,
		path:   resource.URIPath(),
		body:   marshal(resource),
		storeFunc: func(s *mockstore.MockStore) {
			s.On("CreateOrUpdateResource", mock.Anything, mock.AnythingOfType(typ)).
				Return(&store.ErrInternal{}).
				Once()
		},
		wantStatusCode: http.StatusInternalServerError,
	}
}

var updateResourceSuccessTestCase = func(resource corev2.Resource) routerTestCase {
	typ := reflect.TypeOf(resource).String()

	return routerTestCase{
		name:   "it returns 201 if the resource was updated",
		method: http.MethodPut,
		path:   resource.URIPath(),
		body:   marshal(resource),
		storeFunc: func(s *mockstore.MockStore) {
			s.On("CreateOrUpdateResource", mock.Anything, mock.AnythingOfType(typ)).
				Return(nil).
				Once()
		},
		wantStatusCode: http.StatusCreated,
	}
}

// Delete
var deleteTestCases = func(resource corev2.Resource) []routerTestCase {
	return []routerTestCase{
		deleteResourceInvalidPathTestCase(resource),
		deleteResourceNotFoundTestCase(resource),
		deleteResourceStoreErrTestCase(resource),
		deleteResourceSuccessTestCase(resource),
	}
}

var deleteResourceInvalidPathTestCase = func(resource corev2.Resource) routerTestCase {
	return routerTestCase{
		name:           "it returns 400 if the resource ID to delete is invalid",
		method:         http.MethodDelete,
		path:           resource.URIPath() + url.PathEscape("%"),
		wantStatusCode: http.StatusBadRequest,
	}
}

var deleteResourceNotFoundTestCase = func(resource corev2.Resource) routerTestCase {
	return routerTestCase{
		name:   "it returns 404 if the resource to delete does not exist",
		method: http.MethodDelete,
		path:   resource.URIPath(),
		body:   []byte(`{"metadata": {"namespace":"default","name":"foo"}}`),
		storeFunc: func(s *mockstore.MockStore) {
			s.On("DeleteResource", mock.Anything, resource.StorePrefix(), resource.GetObjectMeta().Name).
				Return(&store.ErrNotFound{}).
				Once()
		},
		wantStatusCode: http.StatusNotFound,
	}
}

var deleteResourceStoreErrTestCase = func(resource corev2.Resource) routerTestCase {
	return routerTestCase{
		name:   "it returns 500 if the store returns an error while deleting",
		method: http.MethodDelete,
		path:   resource.URIPath(),
		body:   []byte(`{"metadata": {"namespace":"default","name":"foo"}}`),
		storeFunc: func(s *mockstore.MockStore) {
			s.On("DeleteResource", mock.Anything, resource.StorePrefix(), resource.GetObjectMeta().Name).
				Return(&store.ErrInternal{}).
				Once()
		},
		wantStatusCode: http.StatusInternalServerError,
	}
}

var deleteResourceSuccessTestCase = func(resource corev2.Resource) routerTestCase {
	return routerTestCase{
		name:   "it returns 204 if the resource was deleted",
		method: http.MethodDelete,
		path:   resource.URIPath(),
		body:   []byte(`{"metadata": {"namespace":"default","name":"foo"}}`),
		storeFunc: func(s *mockstore.MockStore) {
			s.On("DeleteResource", mock.Anything, resource.StorePrefix(), resource.GetObjectMeta().Name).
				Return(nil).
				Once()
		},
		wantStatusCode: http.StatusNoContent,
	}
}

func marshal(v interface{}) []byte {
	bytes, _ := json.Marshal(v)
	return bytes
}
