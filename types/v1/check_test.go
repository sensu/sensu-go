package v1

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExtendedAttributesCheckResult(t *testing.T) {
	type getter interface {
		Get(string) (interface{}, error)
	}
	checkResult := CheckResult{
		Name:   "app_01",
		Output: "could not connect to something",
		Client: "proxEnt",
	}
	checkResult.SetExtendedAttributes([]byte(`{"foo":{"bar":42,"baz":9001}}`))
	g, err := checkResult.Get("foo")
	require.NoError(t, err)
	v, err := g.(getter).Get("bar")
	require.NoError(t, err)
	require.Equal(t, 42.0, v)
}
