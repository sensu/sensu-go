package handlers

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/mock"

	corev2 "github.com/sensu/core/v2"
)

func testBody(body string) io.ReadCloser {
	reader := strings.NewReader(body)
	return io.NopCloser(reader)
}

func TestPatchResource(t *testing.T) {
	tests := []struct {
		Name       string
		Body       io.ReadCloser
		StoreError error
		WantErr    bool
		WantUpdate bool
	}{
		{
			Name:       "valid patch",
			Body:       testBody(`{"fake": true}`),
			WantUpdate: true,
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			store := new(mockstore.V2MockStore)
			cs := new(mockstore.ConfigStore)
			store.On("GetConfigStore").Return(cs)
			if test.WantUpdate {
				cs.On("Patch", mock.Anything, mock.Anything, mock.Anything).Return(test.StoreError)
			}
			handlers := Handlers[*corev2.Asset, corev2.Asset]{Store: store}
			req, err := http.NewRequest("PATCH", "http://example.com/namespaces/default/assets/foo", test.Body)
			if err != nil {
				t.Fatal(err)
			}
			_, err = handlers.PatchResource(req)
			if err != nil {
				if !test.WantErr {
					t.Error(err)
				}
			} else if test.WantErr {
				t.Error("wanted non-nil error")
			}
		})
	}
}

func TestValidatePatch(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		vars       map[string]string
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "errors when metadata name & resource id do not match",
			data: []byte(`{"metadata":{"name":"foo"}}`),
			vars: map[string]string{
				"id": "bar",
			},
			wantErr:    true,
			wantErrMsg: "the name of the resource (foo) does not match the name in the URI (bar)",
		},
		{
			name: "errors when metadata namespace & resource namespace do not match",
			data: []byte(`{"metadata":{"namespace":"foo"}}`),
			vars: map[string]string{
				"namespace": "bar",
			},
			wantErr:    true,
			wantErrMsg: "the namespace of the resource (foo) does not match the namespace in the URI (bar)",
		},
		{
			name: "succeeds when resource name & namespace are empty",
			data: []byte(`{}`),
			vars: map[string]string{
				"id":        "foo",
				"namespace": "bar",
			},
			wantErr: false,
		},
		{
			name: "succeeds when name & metadata match resource name & namespace",
			data: []byte(`{"metadata":{"name":"foo","namespace":"bar"}}`),
			vars: map[string]string{
				"id":        "foo",
				"namespace": "bar",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePatch(tt.data, tt.vars)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePatch() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && err.Error() != tt.wantErrMsg {
				t.Errorf("validatePatch() errorMsg = %v, wantErrMsg %v", err, tt.wantErrMsg)
				return
			}
		})
	}
}
