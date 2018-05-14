package e2e

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenSubstitution(t *testing.T) {
	t.Parallel()

	// Initializes sensuctl
	sensuctl, cleanup := newSensuCtl(t)
	defer cleanup()

	// Start the agent
	agentConfig := agentConfig{
		ID:               "TestTokenSubstitution",
		CustomAttributes: `{"team":"devops"}`,
	}
	agent, cleanup := newAgent(agentConfig, sensuctl, t)
	defer cleanup()

	// Create a check that take advantage of token substitution
	check := types.FixtureCheckConfig("check_tokenSubstitution")
	check.Organization = sensuctl.Organization
	check.Environment = sensuctl.Environment
	output, err := sensuctl.run("check", "create", check.Name,
		"--command", `echo {{ .ID }} {{ .Team }} {{ .Missing | default "defaultValue" }}`,
		"--interval", "1",
		"--subscriptions", "test",
		"--handlers", strings.Join(check.Handlers, ","),
		"--organization", check.Organization,
		"--environment", check.Environment,
		"--publish",
	)
	require.NoError(t, err, string(output))

	// Give it few seconds to make sure we've published a check request
	time.Sleep(10 * time.Second)

	output, err = sensuctl.run("event", "info", agent.ID, check.Name)
	require.NoError(t, err, string(output))

	event := types.Event{}
	require.NoError(t, json.Unmarshal(output, &event))
	assert.NotNil(t, event)
	// {{ .ID }} should have been replaced by the entity ID and {{ .Team }} by the
	// custom attribute of the same name on the entity
	assert.Contains(t, event.Check.Output, "TestTokenSubstitution devops defaultValue")

	// Create a check that take advantage of token substitution
	checkUnmatchedToken := types.FixtureCheckConfig("check_unmatchedToken")
	checkUnmatchedToken.Organization = sensuctl.Organization
	checkUnmatchedToken.Environment = sensuctl.Environment
	output, err = sensuctl.run("check", "create", checkUnmatchedToken.Name,
		"--command", "{{ .Foo }}",
		"--interval", "1",
		"--subscriptions", "test",
		"--handlers", strings.Join(check.Handlers, ","),
		"--organization", checkUnmatchedToken.Organization,
		"--environment", checkUnmatchedToken.Environment,
		"--publish",
	)
	require.NoError(t, err, string(output))

	// Give it few seconds to make sure we've published a check request
	time.Sleep(10 * time.Second)

	output, err = sensuctl.run("event", "info", agent.ID, checkUnmatchedToken.Name)
	require.NoError(t, err, string(output))

	event = types.Event{}
	require.NoError(t, json.Unmarshal(output, &event))
	assert.NotNil(t, event)
	// {{ .Foo }} should not have been replaced and an error should have been
	// immediated returned
	assert.Contains(t, event.Check.Output, "has no entry for key")

}
