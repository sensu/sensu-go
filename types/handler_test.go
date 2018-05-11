package types

import (
	"fmt"
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
				Name: "foo",
			},
			Error: "empty handler type",
		},
		{
			Handler: Handler{
				Name: "foo",
				Type: "pipe",
			},
			Error: "environment must be set",
		},
		{
			Handler: Handler{
				Name:        "foo",
				Type:        "pipe",
				Environment: "default",
			},
			Error: "organization must be set",
		},
		{
			Handler: Handler{
				Name:         "foo",
				Type:         "pipe",
				Organization: "default",
				Environment:  "default",
			},
		},
		{
			Handler: Handler{
				Name:         "foo",
				Type:         "grpc",
				Organization: "default",
				Environment:  "default",
			},
		},
		{
			Handler: Handler{
				Name:         "foo",
				Type:         "grpc",
				Organization: "default",
				Environment:  "default",
			},
		},
		{
			Handler: Handler{
				Name:         "foo",
				Type:         "set",
				Organization: "default",
				Environment:  "default",
			},
		},
		{
			Handler: Handler{
				Name:         "foo",
				Type:         "tcp",
				Organization: "default",
				Environment:  "default",
			},
			Error: "tcp and udp handlers need a valid socket",
		},
		{
			Handler: Handler{
				Name:         "foo",
				Type:         "tcp",
				Organization: "default",
				Environment:  "default",
				Socket: &HandlerSocket{
					Host: "localhost",
				},
			},
			Error: "socket port undefined",
		},
		{
			Handler: Handler{
				Name:         "foo",
				Type:         "tcp",
				Organization: "default",
				Environment:  "default",
				Socket: &HandlerSocket{
					Port: 1234,
				},
			},
			Error: "socket host undefined",
		},
		{
			Handler: Handler{
				Name:         "foo",
				Type:         "tcp",
				Organization: "default",
				Environment:  "default",
				Socket: &HandlerSocket{
					Host: "localhost",
					Port: 1234,
				},
			},
		},
		{
			Handler: Handler{
				Name:         "foo",
				Type:         "udp",
				Organization: "default",
				Environment:  "default",
				Socket: &HandlerSocket{
					Host: "localhost",
				},
			},
			Error: "socket port undefined",
		},
		{
			Handler: Handler{
				Name:         "foo",
				Type:         "udp",
				Organization: "default",
				Environment:  "default",
				Socket: &HandlerSocket{
					Port: 1234,
				},
			},
			Error: "socket host undefined",
		},
		{
			Handler: Handler{
				Name:         "foo",
				Type:         "udp",
				Organization: "default",
				Environment:  "default",
				Socket: &HandlerSocket{
					Host: "localhost",
					Port: 1234,
				},
			},
		},
		{
			Handler: Handler{
				Name:         "foo",
				Type:         "magic",
				Organization: "default",
				Environment:  "default",
			},
			Error: "unknown handler type: magic",
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%v", test.Handler), func(t *testing.T) {
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
