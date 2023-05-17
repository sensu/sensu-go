//go:build integration && !race
// +build integration,!race

package etcd

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sensu/sensu-go/backend/store"
)

func TestSessionStore(t *testing.T) {
	testWithEtcd(t, func(s store.Store) {
		ctx := context.Background()
		username := "user1"
		sessionID := "abcdefghijklmnopqrstuvwxyz"
		state := "some stored state"

		// Create some new session state
		err := s.UpdateSession(ctx, username, sessionID, state)
		assert.NoError(t, err)

		// Retrieve session
		retrievedState, err := s.GetSession(ctx, username, sessionID)
		assert.NoError(t, err)
		assert.Equal(t, retrievedState, state)

		// Delete session
		err = s.DeleteSession(ctx, username, sessionID)
		assert.NoError(t, err)

		// Retrieve non existent session
		unknownUser := "unknown_user"
		unknownSessionID := "unknown_session_id"
		_, err = s.GetSession(ctx, unknownUser, unknownSessionID)
		assert.Error(t, err)
	})
}
