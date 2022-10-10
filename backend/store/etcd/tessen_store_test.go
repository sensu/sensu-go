//go:build integration && !race
// +build integration,!race

package etcd

import (
	"context"
	"testing"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/stretchr/testify/assert"
)

func TestTessenStorage(t *testing.T) {
	testWithEtcd(t, func(store store.Store) {
		ctx := context.Background()

		// We should receive an empty config if no tessen config exists
		config, err := store.GetTessenConfig(ctx)
		assert.Error(t, err)
		assert.Empty(t, config)

		// We should not receive an error if the tessen config is valid
		config = &corev2.TessenConfig{
			OptOut: true,
		}
		err = store.CreateOrUpdateTessenConfig(ctx, config)
		assert.NoError(t, err)

		// We should receive an non-empty config if a tessen config exists
		config, err = store.GetTessenConfig(ctx)
		assert.NoError(t, err)
		assert.NotEmpty(t, config)
		assert.True(t, config.OptOut)

		// We should not receive an error if the tessen config is valid
		config = &corev2.TessenConfig{}
		err = store.CreateOrUpdateTessenConfig(ctx, config)
		assert.NoError(t, err)

		// We should receive an empty config if the tessen opt-out is false
		config, err = store.GetTessenConfig(ctx)
		assert.NoError(t, err)
		assert.Empty(t, config)
	})
}
