package graphql

import (
	"testing"

	"github.com/sensu/sensu-go/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func addHook(fns ...InitHook) func() {
	initialHooks := make([]InitHook, 0, len(InitHooks))
	copy(initialHooks, InitHooks)
	InitHooks = append(InitHooks, fns...)
	return func() {
		copy(InitHooks, initialHooks)
	}
}

func TestNewService(t *testing.T) {
	svc, err := NewService(ServiceConfig{})
	require.NoError(t, err)
	assert.NotEmpty(t, svc)
}

func TestInitHooks(t *testing.T) {
	flag := false
	rollback := addHook(func(svc *graphql.Service, cfg ServiceConfig) {
		flag = true
	})
	defer rollback()

	_, err := NewService(ServiceConfig{})
	require.NoError(t, err)
	assert.True(t, flag)
}
