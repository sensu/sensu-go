package routers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/mock"
)

type storeFunc func(*mockstore.V2MockStore)

type routerTestCase struct {
	name           string
	method         string
	path           string
	body           []byte
	storeFunc      storeFunc
	wantStatusCode int
}

func clone(r corev3.Resource) corev3.Resource {
	wrapper, err := storev2.WrapResource(r)
	if err != nil {
		panic(err)
	}
	resource, err := wrapper.Unwrap()
	if err != nil {
		panic(err)
	}
	return resource
}

func run(t *testing.T, tt routerTestCase, router *mux.Router, store *mockstore.V2MockStore) bool {
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
func getTestCases(resource corev3.Resource) []routerTestCase {
	return []routerTestCase{
		getResourceNotFoundTestCase(resource),
		getResourceStoreErrTestCase(resource),
		getResourceSuccessTestCase(resource),
	}
}

func getResourceNotFoundTestCase(resource corev3.Resource) routerTestCase {
	return routerTestCase{
		name:   "it returns 404 if a resource is not found",
		method: http.MethodGet,
		path:   resource.URIPath(),
		storeFunc: func(s *mockstore.V2MockStore) {
			s.On("Get", mock.Anything, mock.Anything).
				Return(nil, &store.ErrNotFound{}).
				Once()
		},
		wantStatusCode: http.StatusNotFound,
	}
}

func getResourceStoreErrTestCase(resource corev3.Resource) routerTestCase {
	return routerTestCase{
		name:   "it returns 500 if the store encounters an error while retrieving a resource",
		method: http.MethodGet,
		path:   resource.URIPath(),
		storeFunc: func(s *mockstore.V2MockStore) {
			s.On("Get", mock.Anything, mock.Anything).
				Return(nil, &store.ErrInternal{}).
				Once()
		},
		wantStatusCode: http.StatusInternalServerError,
	}
}

func getResourceSuccessTestCase(resource corev3.Resource) routerTestCase {
	return routerTestCase{
		name:   "it retrieves a check config",
		method: http.MethodGet,
		path:   resource.URIPath(),
		storeFunc: func(s *mockstore.V2MockStore) {
			s.On("Get", mock.Anything, mock.Anything).
				Return(mockstore.Wrapper[*corev2.CheckConfig]{Value: corev2.FixtureCheckConfig("check")}, nil).
				Once()
		},
		wantStatusCode: http.StatusOK,
	}
}

// List
func listTestCases(resource corev3.Resource) []routerTestCase {
	return []routerTestCase{
		listResourcesStoreErrTestCase(resource),
		listResourcesSuccessTestCase(resource),
	}
}

func listResourcesStoreErrTestCase(resource corev3.Resource) routerTestCase {
	if _, ok := resource.(corev3.GlobalResource); !ok {
		resource.GetMetadata().Namespace = "default"
	}

	return routerTestCase{
		name:   "it returns 500 if the store encounters an error while listing resources",
		method: http.MethodGet,
		path:   resource.URIPath(),
		storeFunc: func(s *mockstore.V2MockStore) {
			s.On("List", mock.Anything, mock.Anything, mock.Anything).
				Return(nil, &store.ErrInternal{}).
				Once()
		},
		wantStatusCode: http.StatusInternalServerError,
	}
}

func listResourcesSuccessTestCase(resource corev3.Resource) routerTestCase {
	if _, ok := resource.(corev3.GlobalResource); !ok {
		resource.GetMetadata().Namespace = "default"
	}

	return routerTestCase{
		name:   "it lists resources",
		method: http.MethodGet,
		path:   resource.URIPath(),
		storeFunc: func(s *mockstore.V2MockStore) {
			s.On("List", mock.Anything, mock.Anything, mock.Anything).
				Return(mockstore.WrapList[*corev2.CheckConfig]{corev2.FixtureCheckConfig("check")}, nil).
				Once()
		},
		wantStatusCode: http.StatusOK,
	}
}

// Create
func createTestCases(resource corev3.Resource) []routerTestCase {
	return []routerTestCase{
		createResourceInvalidPayloadTestCase(resource),
		createResourceInvalidMetaTestCase(resource),
		createResourceAlreadyExistsTestCase(resource),
		createResourceInvalidTestCase(resource),
		createResourceStoreErrTestCase(resource),
		createResourceSuccessTestCase(resource),
	}
}

func createResourceInvalidPayloadTestCase(resource corev3.Resource) routerTestCase {
	resource = clone(resource)

	if _, ok := resource.(corev3.GlobalResource); !ok {
		resource.GetMetadata().Namespace = "default"
	}
	resource.GetMetadata().Name = ""

	return routerTestCase{
		name:           "it returns 400 if the request payload to create is invalid",
		method:         http.MethodPost,
		path:           resource.URIPath(),
		body:           []byte("foo"),
		wantStatusCode: http.StatusBadRequest,
	}
}

func createResourceInvalidMetaTestCase(resource corev3.Resource) routerTestCase {
	resource = clone(resource)
	if _, ok := resource.(corev3.GlobalResource); !ok {
		resource.GetMetadata().Namespace = "default"
	}

	// Do not test this on global resources because it will fail
	if !strings.Contains(resource.URIPath(), "namespaces/default") {
		return routerTestCase{wantStatusCode: 404}
	}

	return routerTestCase{
		name:           "it returns 400 if the resource metadata to create does not match the request path",
		method:         http.MethodPost,
		path:           path.Dir(resource.URIPath()),
		body:           []byte(`{"metadata": {"namespace":"acme"}}`),
		wantStatusCode: http.StatusBadRequest,
	}
}

func createResourceAlreadyExistsTestCase(resource corev3.Resource) routerTestCase {
	// Deep copy the given resource so we can modify it without affecting other
	// test cases
	r := clone(resource)
	namespace := "default"
	if _, ok := r.(corev3.GlobalResource); ok {
		namespace = ""
	}
	r.SetMetadata(&corev2.ObjectMeta{Name: "createResourceAlreadyExistsTestCase", Namespace: namespace})

	body := marshal(r)

	return routerTestCase{
		name:   "it returns 409 if the resource to create already exists",
		method: http.MethodPost,
		path:   path.Dir(resource.URIPath()),
		body:   body,
		storeFunc: func(s *mockstore.V2MockStore) {
			s.On("CreateIfNotExists", mock.Anything, mock.Anything, mock.Anything).
				Return(&store.ErrAlreadyExists{}).
				Once()
		},
		wantStatusCode: http.StatusConflict,
	}
}

func createResourceInvalidTestCase(resource corev3.Resource) routerTestCase {
	// Deep copy the given resource so we can modify it without affecting other
	// test cases
	r := clone(resource)
	namespace := "default"
	if _, ok := r.(corev3.GlobalResource); ok {
		namespace = ""
	}
	r.SetMetadata(&corev2.ObjectMeta{Name: "createResourceInvalidTestCase", Namespace: namespace})

	body := marshal(r)

	return routerTestCase{
		name:   "it returns 400 if the resource to create is invalid",
		method: http.MethodPost,
		path:   path.Dir(resource.URIPath()),
		body:   body,
		storeFunc: func(s *mockstore.V2MockStore) {
			s.On("CreateIfNotExists", mock.Anything, mock.Anything, mock.Anything).
				Return(&store.ErrNotValid{Err: errors.New("createResourceInvalidTestCase")}).
				Once()
		},
		wantStatusCode: http.StatusBadRequest,
	}
}

func createResourceStoreErrTestCase(resource corev3.Resource) routerTestCase {
	// Deep copy the given resource so we can modify it without affecting other
	// test cases
	r := clone(resource)
	namespace := "default"
	if _, ok := r.(corev3.GlobalResource); ok {
		namespace = ""
	}
	r.SetMetadata(&corev2.ObjectMeta{Name: "createResourceStoreErrTestCase", Namespace: namespace})

	body := marshal(r)

	return routerTestCase{
		name:   "it returns 500 if the store returns an error while creating",
		method: http.MethodPost,
		path:   path.Dir(resource.URIPath()),
		body:   body,
		storeFunc: func(s *mockstore.V2MockStore) {
			s.On("CreateIfNotExists", mock.Anything, mock.Anything, mock.Anything).
				Return(&store.ErrInternal{}).
				Once()
		},
		wantStatusCode: http.StatusInternalServerError,
	}
}

func createResourceSuccessTestCase(resource corev3.Resource) routerTestCase {
	// Deep copy the given resource so we can modify it without affecting other
	// test cases
	r := clone(resource)
	namespace := "default"
	if _, ok := r.(corev3.GlobalResource); ok {
		namespace = ""
	}
	r.SetMetadata(&corev2.ObjectMeta{Name: "createResourceSuccessTestCase", Namespace: namespace})

	body := marshal(r)

	return routerTestCase{
		name:   "it returns 201 if the resource was created",
		method: http.MethodPost,
		path:   path.Dir(resource.URIPath()),
		body:   body,
		storeFunc: func(s *mockstore.V2MockStore) {
			s.On("CreateIfNotExists", mock.Anything, mock.Anything, mock.Anything).
				Return(nil).
				Once()
		},
		wantStatusCode: http.StatusCreated,
	}
}

// Update
func updateTestCases(resource corev3.Resource) []routerTestCase {
	return []routerTestCase{
		updateResourceInvalidPayloadTestCase(resource),
		updateResourceInvalidMetaTestCase(resource),
		updateResourceInvalidTestCase(resource),
		updateResourceStoreErrTestCase(resource),
		updateResourceSuccessTestCase(resource),
	}
}

func updateResourceInvalidPayloadTestCase(resource corev3.Resource) routerTestCase {
	return routerTestCase{
		name:           "it returns 400 if the request payload to update is invalid",
		method:         http.MethodPut,
		path:           resource.URIPath(),
		body:           []byte("foo"),
		wantStatusCode: http.StatusBadRequest,
	}
}

func updateResourceInvalidMetaTestCase(resource corev3.Resource) routerTestCase {
	if _, ok := resource.(corev3.GlobalResource); !ok {
		resource.GetMetadata().Namespace = "acme"
	}

	return routerTestCase{
		name:           "it returns 400 if the resource metadata to update does not match the request path",
		method:         http.MethodPut,
		path:           resource.URIPath(),
		body:           []byte(`{"metadata":{"name":"bar"}}`),
		wantStatusCode: http.StatusBadRequest,
	}
}

func updateResourceInvalidTestCase(resource corev3.Resource) routerTestCase {

	return routerTestCase{
		name:   "it returns 400 if the resource to update is invalid",
		method: http.MethodPut,
		path:   resource.URIPath(),
		body:   marshal(resource),
		storeFunc: func(s *mockstore.V2MockStore) {
			s.On("CreateOrUpdate", mock.Anything, mock.Anything, mock.Anything).
				Return(&store.ErrNotValid{Err: errors.New("error")}).
				Once()
		},
		wantStatusCode: http.StatusBadRequest,
	}
}

func updateResourceStoreErrTestCase(resource corev3.Resource) routerTestCase {

	return routerTestCase{
		name:   "it returns 500 if the store returns an error while updating",
		method: http.MethodPut,
		path:   resource.URIPath(),
		body:   marshal(resource),
		storeFunc: func(s *mockstore.V2MockStore) {
			s.On("CreateOrUpdate", mock.Anything, mock.Anything, mock.Anything).
				Return(&store.ErrInternal{}).
				Once()
		},
		wantStatusCode: http.StatusInternalServerError,
	}
}

func updateResourceSuccessTestCase(resource corev3.Resource) routerTestCase {

	return routerTestCase{
		name:   "it returns 201 if the resource was updated",
		method: http.MethodPut,
		path:   resource.URIPath(),
		body:   marshal(resource),
		storeFunc: func(s *mockstore.V2MockStore) {
			s.On("CreateOrUpdate", mock.Anything, mock.Anything, mock.Anything).
				Return(nil).
				Once()
		},
		wantStatusCode: http.StatusCreated,
	}
}

// Delete
func deleteTestCases(resource corev3.Resource) []routerTestCase {
	return []routerTestCase{
		deleteResourceInvalidPathTestCase(resource),
		deleteResourceNotFoundTestCase(resource),
		deleteResourceStoreErrTestCase(resource),
		deleteResourceSuccessTestCase(resource),
	}
}

func deleteResourceInvalidPathTestCase(resource corev3.Resource) routerTestCase {
	return routerTestCase{
		name:           "it returns 400 if the resource ID to delete is invalid",
		method:         http.MethodDelete,
		path:           resource.URIPath() + url.PathEscape("%"),
		wantStatusCode: http.StatusBadRequest,
	}
}

func deleteResourceNotFoundTestCase(resource corev3.Resource) routerTestCase {
	return routerTestCase{
		name:   "it returns 404 if the resource to delete does not exist",
		method: http.MethodDelete,
		path:   resource.URIPath(),
		body:   []byte(`{"metadata": {"namespace":"default","name":"foo"}}`),
		storeFunc: func(s *mockstore.V2MockStore) {
			s.On("Delete", mock.Anything, mock.Anything).
				Return(&store.ErrNotFound{}).
				Once()
			s.On("DeleteNamespace", mock.Anything, mock.Anything).
				Return(&store.ErrNotFound{}).
				Once()
		},
		wantStatusCode: http.StatusNotFound,
	}
}

func deleteResourceStoreErrTestCase(resource corev3.Resource) routerTestCase {
	return routerTestCase{
		name:   "it returns 500 if the store returns an error while deleting",
		method: http.MethodDelete,
		path:   resource.URIPath(),
		body:   []byte(`{"metadata": {"namespace":"default","name":"foo"}}`),
		storeFunc: func(s *mockstore.V2MockStore) {
			s.On("Delete", mock.Anything, mock.Anything).
				Return(&store.ErrInternal{}).
				Once()
			s.On("DeleteNamespace", mock.Anything, mock.Anything).
				Return(&store.ErrInternal{}).
				Once()
		},
		wantStatusCode: http.StatusInternalServerError,
	}
}

func deleteResourceSuccessTestCase(resource corev3.Resource) routerTestCase {
	return routerTestCase{
		name:   "it returns 204 if the resource was deleted",
		method: http.MethodDelete,
		path:   resource.URIPath(),
		body:   []byte(`{"metadata": {"namespace":"default","name":"foo"}}`),
		storeFunc: func(s *mockstore.V2MockStore) {
			s.On("Delete", mock.Anything, mock.Anything).
				Return(nil).
				Once()
			s.On("DeleteNamespace", mock.Anything, mock.Anything).
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
