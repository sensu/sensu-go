package suggest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResourceURIPath(t *testing.T) {
	testCases := []struct {
		in  *Resource
		out string
	}{
		{&Resource{Path: "/api/test"}, "/api/test"},
		{&Resource{Path: "/api/a/b/{namespace}/c"}, "/api/a/b/default/c"},
		{&Resource{Group: "a/b", Name: "c"}, "/api/a/b/c"},
		{&Resource{Group: "core/v2", Name: "c"}, "/api/core/v2/namespaces/default/c"},
	}

	for _, tc := range testCases {
		t.Run(tc.out, func(t *testing.T) {
			url := tc.in.URIPath("default")
			assert.EqualValues(t, tc.out, url)
		})
	}
}

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
