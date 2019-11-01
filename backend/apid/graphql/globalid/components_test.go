package globalid

import (
	"net/url"
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
		{"srn:a:b:c:d", wants{res: "a", nsp: "b", id: "c:d"}},
		{"srn:a:b:c:d?entity=fred", wants{res: "a", nsp: "b", id: "c:d"}},
		{"srn:events:default:x?check=disk&entity=proxy", wants{res: "events", nsp: "default", id: "x"}},
		{"srn:x:y:z", wants{res: "x", nsp: "y", id: "z"}},
	}
	for _, tc := range testCases {
		t.Run(tc.gid, func(t *testing.T) {
			components, err := Parse(tc.gid)
			assert.Equal(tc.want.err, err != nil, "error was expected in result")
			if !tc.want.err {
				assert.Equal(tc.want.res, components.Resource())
				assert.Equal(tc.want.nsp, components.Namespace())
				assert.Equal(tc.want.resType, components.ResourceType())
				assert.Equal(tc.want.id, components.UniqueComponent())
				assert.Equal(tc.gid, components.String())
			}
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
		{
			StandardComponents{
				resource:        "events",
				namespace:       "default",
				uniqueComponent: "xyz",
				extras: url.Values{
					"cat": []string{"dog"},
				},
			},
			"srn:events:default:xyz?cat=dog",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.want, func(t *testing.T) {
			gid := tc.components.String()
			assert.Equal(tc.want, gid)
		})
	}
}

func TestStandardComponents_Extras(t *testing.T) {
	id := &StandardComponents{}
	assert.Empty(t, id.Extras().Get("key"))

	id.Extras().Set("key", "val")
	assert.Equal(t, id.Extras().Get("key"), "val")

	id.Extras().Clear()
	assert.Empty(t, id.Extras().Get("key"))
}
