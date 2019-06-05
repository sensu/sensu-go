package routers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"testing"

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

// Get
var getTestCases = func(pathPrefix, kind string, resource corev2.Resource) []routerTestCase {
	return []routerTestCase{
		getResourceNotFoundTestCase(pathPrefix, kind),
		getResourceStoreErrTestCase(pathPrefix, kind),
		getResourceSuccessTestCase(pathPrefix, kind, resource),
	}
}

var getResourceNotFoundTestCase = func(pathPrefix, kind string) routerTestCase {
	return routerTestCase{
		name:   "it returns 404 if a resource is not found",
		method: http.MethodGet,
		path:   pathPrefix + "/notfound",
		storeFunc: func(s *mockstore.MockStore) {
			s.On("GetResource", mock.Anything, "notfound", mock.AnythingOfType(kind)).
				Return(&store.ErrNotFound{})
		},
		wantStatusCode: http.StatusNotFound,
	}
}

var getResourceStoreErrTestCase = func(pathPrefix, kind string) routerTestCase {
	return routerTestCase{
		name:   "it returns 500 if the store encounters an error while retrieving a resource",
		method: http.MethodGet,
		path:   pathPrefix + "/storerr",
		storeFunc: func(s *mockstore.MockStore) {
			s.On("GetResource", mock.Anything, "storerr", mock.AnythingOfType(kind)).
				Return(&store.ErrInternal{})
		},
		wantStatusCode: http.StatusInternalServerError,
	}
}

var getResourceSuccessTestCase = func(pathPrefix, kind string, resource corev2.Resource) routerTestCase {
	return routerTestCase{
		name:   "it retrieves a check config",
		method: http.MethodGet,
		path:   pathPrefix + "/checkfound",
		storeFunc: func(s *mockstore.MockStore) {
			s.On("GetResource", mock.Anything, "checkfound", mock.AnythingOfType(kind)).
				Return(nil).
				Run(func(args mock.Arguments) {
					args[2] = resource
				})
		},
		wantStatusCode: http.StatusOK,
	}
}

// List
var listTestCases = func(pathPrefix, kind string, resources []corev2.Resource) []routerTestCase {
	return []routerTestCase{
		listResourcesStoreErrTestCase(pathPrefix, kind),
		listResourcesSuccessTestCase(pathPrefix, kind, resources),
		listResourcesAcrossNamespacesTestCase(pathPrefix, kind, resources),
	}
}

var listResourcesStoreErrTestCase = func(pathPrefix, kind string) routerTestCase {
	return routerTestCase{
		name:   "it returns 500 if the store encounters an error while listing resources",
		method: http.MethodGet,
		path:   pathPrefix,
		storeFunc: func(s *mockstore.MockStore) {
			s.On("ListResources", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*[]"+kind), mock.AnythingOfType("*store.SelectionPredicate")).
				Return(&store.ErrInternal{}).
				Once()
		},
		wantStatusCode: http.StatusInternalServerError,
	}
}

var listResourcesSuccessTestCase = func(pathPrefix, kind string, resources []corev2.Resource) routerTestCase {
	return routerTestCase{
		name:   "it lists resources",
		method: http.MethodGet,
		path:   pathPrefix,
		storeFunc: func(s *mockstore.MockStore) {
			s.On("ListResources", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*[]"+kind), mock.AnythingOfType("*store.SelectionPredicate")).
				Return(nil).
				Run(func(args mock.Arguments) {
					args[2] = resources
				}).
				Once()
		},
		wantStatusCode: http.StatusOK,
	}
}

var listResourcesAcrossNamespacesTestCase = func(pathPrefix, kind string, resources []corev2.Resource) routerTestCase {
	// Strip the namespace from the path
	pathPrefix = strings.TrimPrefix(pathPrefix, "/namespaces/default")

	return routerTestCase{
		name:   "it lists resources across namespaces",
		method: http.MethodGet,
		path:   pathPrefix,
		storeFunc: func(s *mockstore.MockStore) {
			s.On("ListResources", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("*[]"+kind), mock.AnythingOfType("*store.SelectionPredicate")).
				Return(nil).
				Run(func(args mock.Arguments) {
					args[2] = resources
				}).
				Once()
		},
		wantStatusCode: http.StatusOK,
	}
}

// Create
var createTestCases = func(pathPrefix, kind string) []routerTestCase {
	return []routerTestCase{
		createResourceInvalidPayloadTestCase(pathPrefix, kind),
		createResourceInvalidMetaTestCase(pathPrefix, kind),
		createResourceAlreadyExistsTestCase(pathPrefix, kind),
		createResourceInvalidTestCase(pathPrefix, kind),
		createResourceStoreErrTestCase(pathPrefix, kind),
		createResourceSuccessTestCase(pathPrefix, kind),
	}
}

var createResourceInvalidPayloadTestCase = func(pathPrefix, kind string) routerTestCase {
	return routerTestCase{
		name:           "it returns 400 if the request payload to create is invalid",
		method:         http.MethodPost,
		path:           pathPrefix,
		body:           []byte("foo"),
		wantStatusCode: http.StatusBadRequest,
	}
}

var createResourceInvalidMetaTestCase = func(pathPrefix, kind string) routerTestCase {
	// Do not test this on global resources
	if s := strings.Split(strings.TrimPrefix(pathPrefix, "/"), "/"); len(s) == 1 {
		return routerTestCase{wantStatusCode: 404}
	}

	return routerTestCase{
		name:           "it returns 400 if the resource metadata to create does not match the request path",
		method:         http.MethodPost,
		path:           pathPrefix,
		body:           []byte(`{"metadata": {"namespace":"acme"}}`),
		wantStatusCode: http.StatusBadRequest,
	}
}

var createResourceAlreadyExistsTestCase = func(pathPrefix, kind string) routerTestCase {
	return routerTestCase{
		name:   "it returns 409 if the resource to create already exists",
		method: http.MethodPost,
		path:   pathPrefix,
		body:   []byte(`{"metadata": {"namespace":"default","name":"foo"}}`),
		storeFunc: func(s *mockstore.MockStore) {
			s.On("CreateResource", mock.Anything, mock.AnythingOfType(kind)).
				Return(&store.ErrAlreadyExists{}).
				Once()
		},
		wantStatusCode: http.StatusConflict,
	}
}

var createResourceInvalidTestCase = func(pathPrefix, kind string) routerTestCase {
	return routerTestCase{
		name:   "it returns 400 if the resource to create is invalid",
		method: http.MethodPost,
		path:   pathPrefix,
		body:   []byte(`{"metadata": {"namespace":"default","name":"foo"}}`),
		storeFunc: func(s *mockstore.MockStore) {
			s.On("CreateResource", mock.Anything, mock.AnythingOfType(kind)).
				Return(&store.ErrNotValid{}).
				Once()
		},
		wantStatusCode: http.StatusBadRequest,
	}
}

var createResourceStoreErrTestCase = func(pathPrefix, kind string) routerTestCase {
	return routerTestCase{
		name:   "it returns 500 if the store returns an error while creating",
		method: http.MethodPost,
		path:   pathPrefix,
		body:   []byte(`{"metadata": {"namespace":"default","name":"foo"}}`),
		storeFunc: func(s *mockstore.MockStore) {
			s.On("CreateResource", mock.Anything, mock.AnythingOfType(kind)).
				Return(&store.ErrInternal{}).
				Once()
		},
		wantStatusCode: http.StatusInternalServerError,
	}
}

var createResourceSuccessTestCase = func(pathPrefix, kind string) routerTestCase {
	return routerTestCase{
		name:   "it returns 204 if the resource was created",
		method: http.MethodPost,
		path:   pathPrefix,
		body:   []byte(`{"metadata": {"namespace":"default","name":"foo"}}`),
		storeFunc: func(s *mockstore.MockStore) {
			s.On("CreateResource", mock.Anything, mock.AnythingOfType(kind)).
				Return(nil).
				Once()
		},
		wantStatusCode: http.StatusNoContent,
	}
}

// Update
var updateTestCases = func(pathPrefix, kind string) []routerTestCase {
	return []routerTestCase{
		updateResourceInvalidPayloadTestCase(pathPrefix, kind),
		updateResourceInvalidMetaTestCase(pathPrefix, kind),
		updateResourceInvalidTestCase(pathPrefix, kind),
		updateResourceStoreErrTestCase(pathPrefix, kind),
		updateResourceSuccessTestCase(pathPrefix, kind),
	}
}

var updateResourceInvalidPayloadTestCase = func(pathPrefix, kind string) routerTestCase {
	return routerTestCase{
		name:           "it returns 400 if the request payload to update is invalid",
		method:         http.MethodPut,
		path:           pathPrefix + "/foo",
		body:           []byte("foo"),
		wantStatusCode: http.StatusBadRequest,
	}
}

var updateResourceInvalidMetaTestCase = func(pathPrefix, kind string) routerTestCase {
	return routerTestCase{
		name:           "it returns 400 if the resource metadata to update does not match the request path",
		method:         http.MethodPut,
		path:           pathPrefix + "/bar",
		body:           []byte(`{"metadata": {"namespace":"default","name":"foo}}`),
		wantStatusCode: http.StatusBadRequest,
	}
}

var updateResourceInvalidTestCase = func(pathPrefix, kind string) routerTestCase {
	return routerTestCase{
		name:   "it returns 400 if the resource to update is invalid",
		method: http.MethodPut,
		path:   pathPrefix + "/foo",
		body:   body(pathPrefix),
		storeFunc: func(s *mockstore.MockStore) {
			s.On("CreateOrUpdateResource", mock.Anything, mock.AnythingOfType(kind)).
				Return(&store.ErrNotValid{}).
				Once()
		},
		wantStatusCode: http.StatusBadRequest,
	}
}

var updateResourceStoreErrTestCase = func(pathPrefix, kind string) routerTestCase {
	return routerTestCase{
		name:   "it returns 500 if the store returns an error while updating",
		method: http.MethodPut,
		path:   pathPrefix + "/foo",
		body:   body(pathPrefix),
		storeFunc: func(s *mockstore.MockStore) {
			s.On("CreateOrUpdateResource", mock.Anything, mock.Anything).
				Return(&store.ErrInternal{}).
				Once()
		},
		wantStatusCode: http.StatusInternalServerError,
	}
}

var updateResourceSuccessTestCase = func(pathPrefix, kind string) routerTestCase {
	return routerTestCase{
		name:   "it returns 204 if the resource was updated",
		method: http.MethodPut,
		path:   pathPrefix + "/foo",
		body:   body(pathPrefix),
		storeFunc: func(s *mockstore.MockStore) {
			s.On("CreateOrUpdateResource", mock.Anything, mock.AnythingOfType(kind)).
				Return(nil).
				Once()
		},
		wantStatusCode: http.StatusNoContent,
	}
}

// Delete
var deleteTestCases = func(pathPrefix, kind string) []routerTestCase {
	return []routerTestCase{
		deleteResourceInvalidPathTestCase(pathPrefix, kind),
		deleteResourceNotFoundTestCase(pathPrefix, kind),
		deleteResourceStoreErrTestCase(pathPrefix, kind),
		deleteResourceSuccessTestCase(pathPrefix, kind),
	}
}

var deleteResourceInvalidPathTestCase = func(pathPrefix, kind string) routerTestCase {
	return routerTestCase{
		name:           "it returns 400 if the resource ID to delete is invalid",
		method:         http.MethodDelete,
		path:           pathPrefix + "/" + url.PathEscape("%"),
		wantStatusCode: http.StatusBadRequest,
	}
}

var deleteResourceNotFoundTestCase = func(pathPrefix, kind string) routerTestCase {
	return routerTestCase{
		name:   "it returns 404 if the resource to delete does not exist",
		method: http.MethodDelete,
		path:   pathPrefix + "/foo",
		body:   []byte(`{"metadata": {"namespace":"default","name":"foo"}}`),
		storeFunc: func(s *mockstore.MockStore) {
			s.On("DeleteResource", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
				Return(&store.ErrNotFound{}).
				Once()
		},
		wantStatusCode: http.StatusNotFound,
	}
}

var deleteResourceStoreErrTestCase = func(pathPrefix, kind string) routerTestCase {
	return routerTestCase{
		name:   "it returns 500 if the store returns an error while deleting",
		method: http.MethodDelete,
		path:   pathPrefix + "/foo",
		body:   []byte(`{"metadata": {"namespace":"default","name":"foo"}}`),
		storeFunc: func(s *mockstore.MockStore) {
			s.On("DeleteResource", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
				Return(&store.ErrInternal{}).
				Once()
		},
		wantStatusCode: http.StatusInternalServerError,
	}
}

var deleteResourceSuccessTestCase = func(pathPrefix, kind string) routerTestCase {
	return routerTestCase{
		name:   "it returns 204 if the resource was delete",
		method: http.MethodDelete,
		path:   pathPrefix + "/foo",
		body:   []byte(`{"metadata": {"namespace":"default","name":"foo"}}`),
		storeFunc: func(s *mockstore.MockStore) {
			s.On("DeleteResource", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).
				Return(nil).
				Once()
		},
		wantStatusCode: http.StatusNoContent,
	}
}

func marshal(t *testing.T, v interface{}) []byte {
	t.Helper()
	bytes, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return bytes
}

func body(pathPrefix string) []byte {
	var body []byte
	if pathPrefix == "/namespaces" {
		body = []byte(`{"name":"foo"}`)
	} else {
		body = []byte(`{"metadata": {"namespace":"default","name":"foo"}}`)
	}
	return body
}
