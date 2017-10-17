package globalid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	assert := assert.New(t)
	type wants struct {
		res     string
		org     string
		env     string
		resType string
		id      string
		err     bool
	}
	testCases := []struct {
		gid  string
		want wants
	}{
		{"users", wants{err: true}},
		{"users:cat/123", wants{err: true}},
		{"users:org:env:cat/123", wants{err: true}},
		{"srn:users", wants{err: true}},
		{"srn:users:123", wants{res: "users", id: "123"}},
		{"srn:u:org:1", wants{res: "u", org: "org", id: "1"}},
		{"srn:users:cat/123", wants{res: "users", resType: "cat", id: "123"}},
	}
	for _, tc := range testCases {
		t.Run(tc.gid, func(t *testing.T) {
			components, err := Parse(tc.gid)
			assert.Equal(tc.want.err, err != nil, "error was expected in result")
			assert.Equal(tc.want.res, components.Resource())
			assert.Equal(tc.want.org, components.Organization())
			assert.Equal(tc.want.env, components.Environment())
			assert.Equal(tc.want.resType, components.ResourceType())
			assert.Equal(tc.want.id, components.UniqueComponent())
		})
	}
}

func TestStandardComponentsString(t *testing.T) {
	assert := assert.New(t)
	testCases := []struct {
		components StandardComponents
		want       string
	}{
		{
			components: StandardComponents{
				resource:        "users",
				uniqueComponent: "123",
			},
			want: "srn:users:123",
		},
		{
			StandardComponents{
				resource:        "users",
				resourceType:    "cat",
				uniqueComponent: "123",
			},
			"srn:users:cat/123",
		},
		{
			StandardComponents{
				resource:        "users",
				organization:    "default",
				uniqueComponent: "123",
			},
			"srn:users:default:123",
		},
		{
			StandardComponents{
				resource:        "users",
				organization:    "default",
				environment:     "default",
				uniqueComponent: "123",
			},
			"srn:users:default:default:123",
		},
		{
			StandardComponents{
				resource:        "users",
				resourceType:    "cat",
				organization:    "default",
				uniqueComponent: "123",
			},
			"srn:users:default:cat/123",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.want, func(t *testing.T) {
			gid := tc.components.String()
			assert.Equal(tc.want, gid)
		})
	}
}
