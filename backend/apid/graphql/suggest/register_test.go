package suggest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterLookup(t *testing.T) {
	reg := Register{
		&Resource{Group: "a/b", Name: "c"},
		&Resource{Group: "b/c", Name: "d"},
		&Resource{Group: "c/d", Name: "e"},
	}

	testCases := []struct {
		ref string
		out *Resource
	}{
		{
			ref: "b/c/d/name",
			out: &Resource{Group: "b/c", Name: "d"},
		},
		{
			ref: "x/y/z/namespace",
			out: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.ref, func(t *testing.T) {
			ref, _ := ParseRef(tc.ref)
			res := reg.Lookup(ref)
			assert.EqualValues(t, tc.out, res)
		})
	}
}
