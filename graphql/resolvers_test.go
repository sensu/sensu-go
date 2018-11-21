package graphql

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultResolver(t *testing.T) {
	type meta struct {
		Name string `json:"firstname"`
	}

	type animal struct {
		meta
		Age int
	}

	fren := animal{meta: meta{Name: "bob"}, Age: 10}

	testCases := []struct {
		desc   string
		source interface{}
		field  string
		out    interface{}
	}{
		{
			desc:   "field on struct",
			source: fren,
			field:  "name",
			out:    "bob",
		},
		{
			desc:   "second field on struct",
			source: fren,
			field:  "age",
			out:    10,
		},
		{
			desc:   "field on struct w/ tag",
			source: fren,
			field:  "firstname",
			out:    "bob",
		},
		{
			desc:   "missing field on struct",
			source: fren,
			field:  "surname",
			out:    nil,
		},
		{
			desc:   "field on map",
			source: map[string]interface{}{"name": "bob"},
			field:  "name",
			out:    "bob",
		},
		{
			desc:   "missing field on map",
			source: map[string]interface{}{"name": "bob"},
			field:  "firstname",
			out:    nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			result, err := DefaultResolver(tc.source, tc.field)
			require.NoError(t, err)
			assert.EqualValues(t, result, tc.out)
		})
	}
}
