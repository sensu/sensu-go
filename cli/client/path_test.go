package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateBasePath(t *testing.T) {
	testCases := []struct {
		ins       []string
		paths     []string
		namespace string
		out       string
	}{
		{
			ins:       []string{"core", "v1", "checks"},
			paths:     []string{},
			namespace: "sensu",
			out:       "/api/core/v1/namespaces/sensu/checks",
		},
		{
			ins:   []string{"core", "v1", "checks"},
			paths: []string{},
			out:   "/api/core/v1/checks",
		},
		{
			ins:       []string{"enterprise", "v7", "distributed-ledgers"},
			paths:     []string{"😂"},
			namespace: "sensu-devel",
			out:       "/api/enterprise/v7/namespaces/sensu-devel/distributed-ledgers/%F0%9F%98%82",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.out, func(t *testing.T) {
			fn := createBasePath(tc.ins[0], tc.ins[1], tc.ins[2])
			out := fn(tc.namespace, tc.paths...)
			assert.Equal(t, tc.out, out)
		})
	}
}
