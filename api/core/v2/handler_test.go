package v2

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFixtureHandler(t *testing.T) {
	handler := FixtureHandler("handler")
	assert.Equal(t, "handler", handler.Name)
	assert.NoError(t, handler.Validate())
}

func TestFixtureSetHandler(t *testing.T) {
	handler := FixtureSetHandler("handler")
	assert.Equal(t, "handler", handler.Name)
	assert.NoError(t, handler.Validate())
}

func TestFixtureSocketHandler(t *testing.T) {
	handler := FixtureSocketHandler("handler", "tcp")
	assert.Equal(t, "handler", handler.Name)
	assert.Equal(t, "tcp", handler.Type)
	assert.NotNil(t, handler.Socket.Host)
	assert.NotNil(t, handler.Socket.Port)
	assert.NoError(t, handler.Validate())
}

func TestHandlerValidate(t *testing.T) {
	tests := []struct {
		Handler Handler
		Error   string
	}{
		{
			Handler: Handler{},
			Error:   "handler name must not be empty",
		},
		{
			Handler: Handler{
				ObjectMeta: ObjectMeta{
					Name: "foo",
				},
			},
			Error: "empty handler type",
		},
		{
			Handler: Handler{
				ObjectMeta: ObjectMeta{
					Name: "foo",
				},
				Type:    "pipe",
				Command: "sl",
			},
			Error: "namespace must be set",
		},
		{
			Handler: Handler{
				ObjectMeta: ObjectMeta{
					Name:      "foo",
					Namespace: "default",
				},
				Type:    "pipe",
				Command: "sl",
			},
		},
		{
			Handler: Handler{
				ObjectMeta: ObjectMeta{
					Name:      "foo",
					Namespace: "default",
				},
				Type: "grpc",
			},
		},
		{
			Handler: Handler{
				ObjectMeta: ObjectMeta{
					Name:      "foo",
					Namespace: "default",
				},
				Type: "grpc",
			},
		},
		{
			Handler: Handler{
				ObjectMeta: ObjectMeta{
					Name:      "foo",
					Namespace: "default",
				},
				Type: "set",
			},
		},
		{
			Handler: Handler{
				ObjectMeta: ObjectMeta{
					Name:      "foo",
					Namespace: "default",
				},
				Type: "tcp",
			},
			Error: "tcp and udp handlers need a valid socket",
		},
		{
			Handler: Handler{
				ObjectMeta: ObjectMeta{
					Name:      "foo",
					Namespace: "default",
				},
				Type: "tcp",
				Socket: &HandlerSocket{
					Host: "localhost",
				},
			},
			Error: "socket port undefined",
		},
		{
			Handler: Handler{
				ObjectMeta: ObjectMeta{
					Name:      "foo",
					Namespace: "default",
				},
				Type: "tcp",
				Socket: &HandlerSocket{
					Port: 1234,
				},
			},
			Error: "socket host undefined",
		},
		{
			Handler: Handler{
				ObjectMeta: ObjectMeta{
					Name:      "foo",
					Namespace: "default",
				},
				Type: "tcp",
				Socket: &HandlerSocket{
					Host: "localhost",
					Port: 1234,
				},
			},
		},
		{
			Handler: Handler{
				ObjectMeta: ObjectMeta{
					Name:      "foo",
					Namespace: "default",
				},
				Type: "udp",
				Socket: &HandlerSocket{
					Host: "localhost",
				},
			},
			Error: "socket port undefined",
		},
		{
			Handler: Handler{
				ObjectMeta: ObjectMeta{
					Name:      "foo",
					Namespace: "default",
				},
				Type: "udp",
				Socket: &HandlerSocket{
					Port: 1234,
				},
			},
			Error: "socket host undefined",
		},
		{
			Handler: Handler{
				ObjectMeta: ObjectMeta{
					Name:      "foo",
					Namespace: "default",
				},
				Type: "udp",
				Socket: &HandlerSocket{
					Host: "localhost",
					Port: 1234,
				},
			},
		},
		{
			Handler: Handler{
				ObjectMeta: ObjectMeta{
					Name:      "foo",
					Namespace: "default",
				},
				Type: "magic",
			},
			Error: "unknown handler type: magic",
		},
		{
			Handler: Handler{
				ObjectMeta: ObjectMeta{
					Name: "foo",
				},
				Type: "pipe",
			},
			Error: "missing command",
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("%d", i), func(t *testing.T) {
			if err := test.Handler.Validate(); err != nil {
				if len(test.Error) > 0 {
					require.Equal(t, test.Error, err.Error())
				} else {
					t.Fatal(err)
				}
			} else if len(test.Error) > 0 {
				t.Fatal("expected error, got none")
			}
		})
	}
}

func TestSortHandlersByName(t *testing.T) {
	a := FixtureHandler("Abernathy")
	b := FixtureHandler("Bernard")
	c := FixtureHandler("Clementine")
	d := FixtureHandler("Dolores")

	testCases := []struct {
		name     string
		inDir    bool
		inChecks []*Handler
		expected []*Handler
	}{
		{
			name:     "Sorts ascending",
			inDir:    true,
			inChecks: []*Handler{d, c, a, b},
			expected: []*Handler{a, b, c, d},
		},
		{
			name:     "Sorts descending",
			inDir:    false,
			inChecks: []*Handler{d, a, c, b},
			expected: []*Handler{d, c, b, a},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			sort.Sort(SortHandlersByName(tc.inChecks, tc.inDir))
			assert.EqualValues(t, tc.expected, tc.inChecks)
		})
	}
}
