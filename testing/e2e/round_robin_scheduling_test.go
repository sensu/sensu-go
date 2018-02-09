package e2e

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoundRobinScheduling(t *testing.T) {
	t.Parallel()

	// Create two backends
	backendA, cleanup := newBackend(t)
	defer cleanup()

	sensuctlA, cleanup := newSensuCtl(backendA.HTTPURL, "default", "default", "admin", "P@ssw0rd!")
	defer cleanup()

	// TODO(echlebek): make this test work with multiple backends
	//backendB, cleanup := newBackend(t)
	//defer cleanup()

	sensuctlB, cleanup := newSensuCtl(backendA.HTTPURL, "default", "default", "admin", "P@ssw0rd!")
	defer cleanup()

	// Two agents belong to backend A, one belongs to backend B
	agentCfgA := agentConfig{
		ID:          "agentA",
		BackendURLs: []string{backendA.WSURL},
	}
	agentCfgB := agentConfig{
		ID:          "agentB",
		BackendURLs: []string{backendA.WSURL},
	}
	agentCfgC := agentConfig{
		ID:          "agentC",
		BackendURLs: []string{backendA.WSURL},
	}

	agentA, cleanup := newAgent(agentCfgA, sensuctlA, t)
	defer cleanup()

	agentB, cleanup := newAgent(agentCfgB, sensuctlA, t)
	defer cleanup()

	agentC, cleanup := newAgent(agentCfgC, sensuctlB, t)
	defer cleanup()

	// Create an authenticated HTTP Sensu client. newSensuClient is deprecated but
	// sensuctl does not currently support objects updates with flag parameters
	clientA := newSensuClient(backendA.HTTPURL)
	clientB := newSensuClient(backendA.HTTPURL)
	clientC := newSensuClient(backendA.HTTPURL)

	// Create a check that publish check requests
	check := types.FixtureCheckConfig("TestCheckScheduling")
	check.Publish = true
	check.Interval = 1
	check.Subscriptions = []string{"test"}
	check.RoundRobin = true
	check.Command = testutil.CommandPath(filepath.Join(toolsDir, "true"))

	err := clientA.CreateCheck(check)
	require.NoError(t, err)
	check, err = clientA.FetchCheck(check.Name)
	require.NoError(t, err)

	// Allow checks to be published
	time.Sleep(20 * time.Second)

	eventA, err := clientA.FetchEvent(agentA.ID, check.Name)
	require.NoError(t, err)
	require.NotNil(t, eventA)

	eventB, err := clientB.FetchEvent(agentB.ID, check.Name)
	require.NoError(t, err)
	require.NotNil(t, eventB)

	eventC, err := clientC.FetchEvent(agentC.ID, check.Name)
	require.NoError(t, err)
	require.NotNil(t, eventC)

	histories := append(eventA.Check.History, eventB.Check.History...)
	histories = append(histories, eventC.Check.History...)

	executed := make(map[int64]struct{})
	for _, h := range histories {
		assert.Equal(t, int32(0), h.Status)
		e := h.Executed
		executed[e] = struct{}{}
	}
	// Ensure that all executed checks have been executed at a separate time.
	// TODO(echlebek): do this with unique identifiers per check request msg.
	assert.Equal(t, len(histories), len(executed))
}
