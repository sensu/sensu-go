package api

import (
	"reflect"
	"testing"
)

func TestParseVersion(t *testing.T) {
	tests := []struct {
		Input  string
		Output Version
	}{
		{
			Input: "v1",
			Output: Version{
				Number: 1,
			},
		},
		{
			Input: "v1alpha1",
			Output: Version{
				Number:        1,
				Interim:       "alpha",
				InterimNumber: 1,
			},
		},
	}

	for _, test := range tests {
		v, err := ParseVersion(test.Input)
		if err != nil {
			t.Fatal(err)
		}
		if got, want := v, test.Output; !reflect.DeepEqual(got, want) {
			t.Fatalf("bad version: got %v, want %v", got, want)
		}
	}
}
