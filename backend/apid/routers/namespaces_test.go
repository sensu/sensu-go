package routers

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/api/mockapi"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/stretchr/testify/mock"
)

type mockPatcher struct {
	mock.Mock
}

func (m *mockPatcher) PatchResource(req *http.Request) (corev3.Resource, error) {
	args := m.Called(req)
	return args.Get(0).(corev3.Resource), args.Error(1)
}

func TestNamespacesRouter(t *testing.T) {
	mockctl := gomock.NewController(t)
	defer mockctl.Finish()

	nsClient := mockapi.NewMockNamespaceClient(mockctl)
	//nsClient.EXPECT().ListNamespaces(gomock.Any(), gomock.Any()).Return([]*corev3.Namespace{corev3.FixtureNamespace("default")}, nil)
	//nsClient.EXPECT().CreateNamespace(gomock.Any(), gomock.Any()).Return(nil)
	//nsClient.EXPECT().UpdateNamespace(gomock.Any(), gomock.Any()).Return(nil)
	//nsClient.EXPECT().DeleteNamespace(gomock.Any(), gomock.Any()).Return(nil)
	nsClient.EXPECT().FetchNamespace(gomock.Any(), "default").Return(corev3.FixtureNamespace("default"), nil)
	nsClient.EXPECT().FetchNamespace(gomock.Any(), "other").Return(nil, &store.ErrNotFound{})
	nsClient.EXPECT().FetchNamespace(gomock.Any(), "broken").Return(nil, errors.New("what"))

	patcher := new(mockPatcher)
	patcher.On("Patch", mock.Anything).Return(corev3.FixtureNamespace("default"), nil)

	router := NewNamespacesRouter(nsClient, patcher)

	tests := []struct {
		Method    string
		Path      string
		Body      []byte
		ExpStatus int
	}{
		{
			Method:    "GET",
			Path:      "/namespaces/default",
			ExpStatus: http.StatusOK,
		},
		{
			Method:    "GET",
			Path:      "/namespaces/other",
			ExpStatus: http.StatusNotFound,
		},
		{
			Method:    "GET",
			Path:      "/namespaces/broken",
			ExpStatus: http.StatusInternalServerError,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s%s", test.Method, test.Path), func(t *testing.T) {
			mux := mux.NewRouter().UseEncodedPath()
			router.Mount(mux)
			server := httptest.NewServer(mux)
			defer server.Close()

			client := new(http.Client)
			req, err := http.NewRequest(test.Method, server.URL+test.Path, bytes.NewReader(test.Body))
			if err != nil {
				t.Fatal(err)
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}

			defer resp.Body.Close()

			if got, want := resp.StatusCode, test.ExpStatus; got != want {
				t.Errorf("bad status: got %d, want %d", got, want)
				body, _ := ioutil.ReadAll(resp.Body)
				t.Error(string(body))
			}
		})
	}
}
