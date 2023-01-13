package suggest

import (
	"sort"
	"testing"

	v2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/stretchr/testify/assert"
)

func TestCustomFieldMatches(t *testing.T) {
	field := CustomField{
		Name: "my-field",
		FieldFunc: func(corev3.Resource) []string {
			return []string{"test"}
		},
	}

	testCases := []struct {
		in  string
		out bool
	}{
		{in: "my-field", out: true},
		{in: "my-field/test", out: false},
		{in: "my-fields", out: false},
	}

	for _, tc := range testCases {
		t.Run(tc.in, func(t *testing.T) {
			out := field.Matches(tc.in)
			assert.EqualValues(t, out, tc.out)
		})
	}
}

func TestCustomFieldValue(t *testing.T) {
	check := v2.FixtureCheckConfig("check")
	field := CustomField{
		Name: "my-field",
		FieldFunc: func(res corev3.Resource) []string {
			return res.(*v2.CheckConfig).Subscriptions
		},
	}

	out := field.Value(check, "---")
	assert.EqualValues(t, out, check.Subscriptions)
}

func TestMapFieldMatches(t *testing.T) {
	field := MapField{
		Name: "a/b",
		FieldFunc: func(corev3.Resource) map[string]string {
			return map[string]string{}
		},
	}

	testCases := []struct {
		in  string
		out bool
	}{
		{in: "a/b", out: true},
		{in: "a/b/", out: true},
		{in: "a/b/c", out: true},
		{in: "a/b/c/d", out: true},
		{in: "a", out: false},
		{in: "b/a", out: false},
	}

	for _, tc := range testCases {
		t.Run(tc.in, func(t *testing.T) {
			out := field.Matches(tc.in)
			assert.EqualValues(t, tc.out, out)
		})
	}
}

func TestMapFieldValues(t *testing.T) {
	check := v2.FixtureCheckConfig("check")
	field := MapField{
		Name: "a/b",
		FieldFunc: func(corev3.Resource) map[string]string {
			return map[string]string{
				"a": "b",
				"c": "d",
				"e": "f",
			}
		},
	}

	testCases := []struct {
		in  string
		out []string
	}{
		{in: "a/b", out: []string{"a", "c", "e"}},
		{in: "a/b/c", out: []string{"d"}},
		{in: "a/b/c/d", out: []string{}},
	}

	for _, tc := range testCases {
		t.Run(tc.in, func(t *testing.T) {
			out := field.Value(check, tc.in)
			sort.Strings(out)
			assert.EqualValues(t, tc.out, out)
		})
	}
}
