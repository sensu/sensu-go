package globalid

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	assert := assert.New(t)
	type wants struct {
		res     string
		nsp     string
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
		{"users:ns:cat/123", wants{err: true}},
		{"srn:users", wants{err: true}},
		{"srn:users:123", wants{res: "users", id: "123"}},
		{"srn:u:ns:1", wants{res: "u", nsp: "ns", id: "1"}},
		{"srn:users:cat/123", wants{res: "users", resType: "cat", id: "123"}},
		{"srn:x:y:*:z", wants{res: "x", nsp: "y", id: "*:z"}},
	}
	for _, tc := range testCases {
		t.Run(tc.gid, func(t *testing.T) {
			components, err := Parse(tc.gid)
			assert.Equal(tc.want.err, err != nil, "error was expected in result")
			assert.Equal(tc.want.res, components.Resource())
			assert.Equal(tc.want.nsp, components.Namespace())
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
				namespace:       "default",
				uniqueComponent: "123",
			},
			"srn:users:default:123",
		},
		{
			StandardComponents{
				resource:        "users",
				namespace:       "default",
				uniqueComponent: "123",
			},
			"srn:users:default:123",
		},
		{
			StandardComponents{
				resource:        "users",
				resourceType:    "cat",
				namespace:       "default",
				uniqueComponent: "123",
			},
			"srn:users:default:cat/123",
		},
		{
			StandardComponents{
				resource:        "silences",
				namespace:       "default",
				uniqueComponent: "123:456",
			},
			"srn:silences:default:123:456",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.want, func(t *testing.T) {
			gid := tc.components.String()
			assert.Equal(tc.want, gid)
		})
	}
}
