package e2e

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/sensu/sensu-go/types"
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
	if err := backoff.Retry(func(retry int) (bool, error) {
		if output, err = sensuctl.run("event", "info", agent.ID, check.Name); err != nil {
			// The command returned an error, let's retry
			return false, nil
		}

		event := &types.Event{}
		if err := json.Unmarshal(output, event); err != nil || event == nil {
			return false, nil
		}

		// {{ .ID }} should have been replaced by the entity ID and {{ .Team }} by
		// the custom attribute of the same name on the entity
		if !strings.Contains(event.Check.Output, "TestTokenSubstitution devops defaultValue") {
			return false, nil
		}

		// At this point the attempt was successful
		return true, nil
	}); err != nil {
		t.Errorf("no check event was received: %s", string(output))
	}

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
	if err := backoff.Retry(func(retry int) (bool, error) {
		if output, err = sensuctl.run("event", "info", agent.ID, checkUnmatchedToken.Name); err != nil {
			// The command returned an error, let's retry
			return false, nil
		}

		event := &types.Event{}
		if err := json.Unmarshal(output, event); err != nil || event == nil {
			return false, nil
		}

		// {{ .Foo }} should not have been replaced and an error should have been
		// // immediated returned
		if !strings.Contains(event.Check.Output, "has no entry for key") {
			return false, nil
		}

		// At this point the attempt was successful
		return true, nil
	}); err != nil {
		t.Errorf("no check event was received: %s", string(output))
	}
}
