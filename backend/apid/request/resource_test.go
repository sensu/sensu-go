package request

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/types"
)

func TestResource(t *testing.T) {
	testCases := []struct {
		Name    string
		Request *http.Request

		ExpectedResource corev3.Resource
		CheckError       func(*testing.T, error)
	}{
		{
			Name:       "empty request errors",
			Request:    newRequest([]byte("{}")),
			CheckError: assertError,
		}, {
			Name: "error for legacy request format",
			Request: newRequest(func() []byte {
				b, _ := json.Marshal(corev2.FixtureCheckConfig("test"))
				return b
			}()),
			CheckError: func(t *testing.T, err error) {
				assertError(t, err)
			},
		}, {
			Name: "error for different resource type",
			Request: newRequest(func() []byte {
				b, _ := json.Marshal(types.WrapResource(corev2.FixtureHandler("test")))
				return b
			}()),
			CheckError: assertError,
		}, {
			Name: "success",
			Request: newRequest(func() []byte {
				b, _ := json.Marshal(types.WrapResource(corev2.FixtureCheckConfig("test")))
				return b
			}()),
			CheckError: func(t *testing.T, err error) {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			actual, err := Resource[*corev2.CheckConfig](tc.Request)
			tc.CheckError(t, err)
			if tc.ExpectedResource == nil {
				return
			}
			if !cmp.Equal(tc.ExpectedResource, actual) {
				t.Error(cmp.Diff(tc.ExpectedResource, actual))
			}
		})
	}
}

func newRequest(b []byte) *http.Request {
	r, _ := http.NewRequest(http.MethodHead, "", bytes.NewBuffer(b))
	return r
}

func assertError(t *testing.T, err error) {
	if err == nil {
		t.Error("expected error, got nil")
	}
}
