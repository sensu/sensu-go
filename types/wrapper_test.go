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
				fmt.Println(string(test.Body))
			}
		})
	}
}
