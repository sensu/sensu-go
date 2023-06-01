package routers

import (
	"context"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gorilla/mux"
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/backend/apid/handlers"
	"github.com/sensu/sensu-go/backend/store"
)

func newRequest(t *testing.T, method, endpoint string, body io.Reader) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, endpoint, body)
	if err != nil {
		t.Fatal(err)
	}
	return req.WithContext(context.Background())
}

func TestRespondWith(t *testing.T) {
	type args struct {
		w        http.ResponseWriter
		r        *http.Request
		response handlers.HandlerResponse
	}
	fixtureEntity := corev2.FixtureEntity("hello")
	fixtureEntity.Annotations = map[string]string{
		store.SensuETagKey: base64.RawStdEncoding.EncodeToString([]byte("helloworld")),
	}
	tests := []struct {
		name             string
		args             args
		expectETagHeader bool
	}{
		{
			name: "etag not added when resource is []corev2.Resource",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("GET", "/", nil),
				response: handlers.HandlerResponse{
					ResourceList: []corev3.Resource{
						corev2.FixtureEntity("entity1"),
					},
				},
			},
			expectETagHeader: false,
		},
		{
			name: "etag added when resource is corev2.Resource",
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("GET", "/", nil),
				response: handlers.HandlerResponse{
					Resource: fixtureEntity,
				},
			},
			expectETagHeader: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RespondWith(tt.args.w, tt.args.r, tt.args.response)
			if tt.expectETagHeader {
				if _, ok := tt.args.w.Header()["ETag"]; !ok {
					t.Errorf("RespondWith() did not set ETag header, headers: %v", tt.args.w.Header())
				}
			}
		})
	}
}

func TestWriteError(t *testing.T) {
	type args struct {
		w   http.ResponseWriter
		err error
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			WriteError(tt.args.w, tt.args.err)
		})
	}
}

func TestHTTPStatusFromCode(t *testing.T) {
	type args struct {
		code actions.ErrCode
	}
	tests := []struct {
		name string
		args args
		want int
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HTTPStatusFromCode(tt.args.code); got != tt.want {
				t.Errorf("HTTPStatusFromCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_actionHandler(t *testing.T) {
	type args struct {
		action actionHandlerFunc
	}
	tests := []struct {
		name string
		args args
		want http.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := actionHandler(tt.args.action); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("actionHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_listHandler(t *testing.T) {
	type args struct {
		fn listHandlerFunc
	}
	tests := []struct {
		name string
		args args
		want http.HandlerFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := listHandler(tt.args.fn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("listHandler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResourceRoute_Get(t *testing.T) {
	type fields struct {
		Router     *mux.Router
		PathPrefix string
	}
	type args struct {
		fn actionHandlerFunc
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *mux.Route
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ResourceRoute{
				Router:     tt.fields.Router,
				PathPrefix: tt.fields.PathPrefix,
			}
			if got := r.Get(tt.args.fn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ResourceRoute.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResourceRoute_List(t *testing.T) {
	type fields struct {
		Router     *mux.Router
		PathPrefix string
	}
	type args struct {
		fn     ListControllerFunc
		fields FieldsFunc
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *mux.Route
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ResourceRoute{
				Router:     tt.fields.Router,
				PathPrefix: tt.fields.PathPrefix,
			}
			if got := r.List(tt.args.fn, tt.args.fields); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ResourceRoute.List() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResourceRoute_ListAllNamespaces(t *testing.T) {
	type fields struct {
		Router     *mux.Router
		PathPrefix string
	}
	type args struct {
		fn     ListControllerFunc
		path   string
		fields FieldsFunc
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *mux.Route
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ResourceRoute{
				Router:     tt.fields.Router,
				PathPrefix: tt.fields.PathPrefix,
			}
			if got := r.ListAllNamespaces(tt.args.fn, tt.args.path, tt.args.fields); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ResourceRoute.ListAllNamespaces() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResourceRoute_Post(t *testing.T) {
	type fields struct {
		Router     *mux.Router
		PathPrefix string
	}
	type args struct {
		fn actionHandlerFunc
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *mux.Route
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ResourceRoute{
				Router:     tt.fields.Router,
				PathPrefix: tt.fields.PathPrefix,
			}
			if got := r.Post(tt.args.fn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ResourceRoute.Post() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResourceRoute_Put(t *testing.T) {
	type fields struct {
		Router     *mux.Router
		PathPrefix string
	}
	type args struct {
		fn actionHandlerFunc
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *mux.Route
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ResourceRoute{
				Router:     tt.fields.Router,
				PathPrefix: tt.fields.PathPrefix,
			}
			if got := r.Put(tt.args.fn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ResourceRoute.Put() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResourceRoute_Del(t *testing.T) {
	type fields struct {
		Router     *mux.Router
		PathPrefix string
	}
	type args struct {
		fn actionHandlerFunc
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *mux.Route
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ResourceRoute{
				Router:     tt.fields.Router,
				PathPrefix: tt.fields.PathPrefix,
			}
			if got := r.Del(tt.args.fn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ResourceRoute.Del() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestResourceRoute_Path(t *testing.T) {
	type fields struct {
		Router     *mux.Router
		PathPrefix string
	}
	type args struct {
		p  string
		fn actionHandlerFunc
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *mux.Route
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &ResourceRoute{
				Router:     tt.fields.Router,
				PathPrefix: tt.fields.PathPrefix,
			}
			if got := r.Path(tt.args.p, tt.args.fn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ResourceRoute.Path() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_handleAction(t *testing.T) {
	type args struct {
		router *mux.Router
		path   string
		fn     actionHandlerFunc
	}
	tests := []struct {
		name string
		args args
		want *mux.Route
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := handleAction(tt.args.router, tt.args.path, tt.args.fn); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("handleAction() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUnmarshalBody(t *testing.T) {
	type args struct {
		req    *http.Request
		record interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
		})
	}
}
