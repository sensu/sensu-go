package v2

import (
	"reflect"
	"testing"
)

func TestValidateEnvVars(t *testing.T) {
	tests := []struct {
		Name     string
		EnvVars  []string
		ExpError bool
	}{
		{
			Name: "empty",
		},
		{
			Name:    "it should work",
			EnvVars: []string{"FOO=BAR", "BAZ=FOOBAR"},
		},
		{
			Name:     "it should not work",
			EnvVars:  []string{"FOO=BAR", "foo:bar"},
			ExpError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			err := ValidateEnvVars(test.EnvVars)
			if test.ExpError && err == nil {
				t.Fatal("expected error")
			}
			if !test.ExpError && err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestEnvVarsToMap(t *testing.T) {
	tests := []struct {
		Name    string
		EnvVars []string
		Exp     map[string]string
	}{
		{
			Name: "empty",
			Exp:  map[string]string{},
		},
		{
			Name:    "no invalid",
			EnvVars: []string{"FOO=BAR", "BAZ=FOOBAR"},
			Exp:     map[string]string{"FOO": "BAR", "BAZ": "FOOBAR"},
		},
		{
			Name:    "some invalid",
			EnvVars: []string{"FOO=BAR", "foo:bar"},
			Exp:     map[string]string{"FOO": "BAR"},
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			m := EnvVarsToMap(test.EnvVars)
			if got, want := m, test.Exp; !reflect.DeepEqual(got, want) {
				t.Fatalf("bad result: got %v, want %v", got, want)
			}
		})
	}
}
