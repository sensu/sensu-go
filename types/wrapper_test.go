package types

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	v2 "github.com/sensu/sensu-go/api/core/v2"
)

var null = json.RawMessage("null")

type generic map[string]*json.RawMessage

func (g generic) URIPath() string {
	return ""
}

func (g generic) Validate() error {
	return nil
}

func (g generic) SetNamespace(namespace string) {
	return
}

func mustMarshal(t *testing.T, value interface{}) []byte {
	t.Helper()

	b, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}

	return b
}

func TestUnmarshalBody(t *testing.T) {
	asset := FixtureAsset("bar")
	asset.Labels["foo"] = "bar"
	var (
		wrappedAsset = Wrapper{
			TypeMeta: corev2.TypeMeta{
				Type:       "Asset",
				APIVersion: "core/v2",
			},
			ObjectMeta: ObjectMeta{
				Labels: map[string]string{
					"bar": "baz",
				},
			},
			Value: FixtureAsset("bar"),
		}
	)

	wrappedAssetB := mustMarshal(t, wrappedAsset)

	tests := []struct {
		Name   string
		Body   []byte
		Value  interface{}
		ExpErr bool
	}{
		{
			Name:  "wrapped regular type, no extras",
			Body:  wrappedAssetB,
			Value: &wrappedAsset,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			err := json.Unmarshal(test.Body, test.Value)
			if err != nil && !test.ExpErr {
				t.Fatal(err)
			}
			if err == nil && test.ExpErr {
				t.Fatal("expected an error")
			}
		})
	}
}

func TestWrapResourceObjectMeta(t *testing.T) {
	check := FixtureCheck("foo")
	check.Labels["asdf"] = "asdf"

	wrapped := WrapResource(check)
	if !reflect.DeepEqual(wrapped.ObjectMeta, check.ObjectMeta) {
		t.Fatal("objectmeta not equal")
	}
}

func TestResolveType(t *testing.T) {
	testCases := []struct {
		ApiVersion string
		Type       string
		ExpRet     interface{}
		ExpErr     bool
	}{
		{
			ApiVersion: "core/v2",
			Type:       "asset",
			ExpRet:     &v2.Asset{},
			ExpErr:     false,
		},
		{
			ApiVersion: "non/existence",
			Type:       "null",
			ExpRet:     nil,
			ExpErr:     true,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s/%s", tc.ApiVersion, tc.Type), func(t *testing.T) {
			r, err := ResolveType(tc.ApiVersion, tc.Type)
			if !reflect.DeepEqual(r, tc.ExpRet) {
				t.Fatal("unexpected type")
			}
			if err != nil && !tc.ExpErr {
				t.Fatal(err)
			}
			if err == nil && tc.ExpErr {
				t.Fatal("expected an error")
			}
		})
	}
}
