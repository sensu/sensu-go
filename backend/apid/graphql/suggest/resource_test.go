package suggest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResourceLookupField(t *testing.T) {
	res := &Resource{
		Fields: []Field{
			&CustomField{Name: "one"},
			&CustomField{Name: "two"},
		},
	}

	testCases := []struct {
		ref string
		out Field
	}{
		{"a/b/c/one", &CustomField{Name: "one"}},
		{"a/b/c/two", &CustomField{Name: "two"}},
		{"a/b/c/three", nil},
	}

	for _, tc := range testCases {
		t.Run(tc.ref, func(t *testing.T) {
			ref, _ := ParseRef(tc.ref)
			f := res.LookupField(ref)
			assert.EqualValues(t, tc.out, f)
		})
	}
}
