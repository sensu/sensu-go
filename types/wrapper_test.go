package types

import (
	"encoding/json"
	"fmt"
	"testing"
)

var null = json.RawMessage("null")

type generic map[string]*json.RawMessage

func (g generic) URIPath() string {
	return ""
}

func (g generic) Validate() error {
	return nil
}

func addExtendedAttrToWrapped(t *testing.T, wrapped Wrapper) []byte {
	t.Helper()

	b := mustMarshal(t, wrapped.Value)
	generic := make(generic)
	if err := json.Unmarshal(b, &generic); err != nil {
		t.Fatal(err)
	}
	generic["unknown"] = &null

	return mustMarshal(t, Wrapper{Type: wrapped.Type, Value: generic})
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
	var (
		wrappedAsset = Wrapper{Type: "Asset", Value: FixtureAsset("bar")}
		wrappedCheck = Wrapper{Type: "Check", Value: FixtureCheck("barbaz")}
	)

	wrappedAssetB := mustMarshal(t, wrappedAsset)
	wrappedCheckB := mustMarshal(t, wrappedCheck)

	badWrappedAssetB := addExtendedAttrToWrapped(t, wrappedAsset)
	goodWrappedCheckB := addExtendedAttrToWrapped(t, wrappedCheck)

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
		{
			Name:  "wrapped extended attributes type, no extras",
			Body:  wrappedCheckB,
			Value: &wrappedCheck,
		},
		{
			Name:   "wrapped type with extras",
			Body:   badWrappedAssetB,
			Value:  &wrappedAsset,
			ExpErr: true,
		},
		{
			Name:  "wrapped extended attributes type, with extras",
			Body:  goodWrappedCheckB,
			Value: &wrappedCheck,
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
				fmt.Println(string(test.Body))
			}
		})
	}
}
