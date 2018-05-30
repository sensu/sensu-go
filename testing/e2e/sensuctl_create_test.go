package e2e

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSensuctlCreate(t *testing.T) {
	t.Parallel()

	// Initializes sensuctl
	sensuctl, cleanup := newSensuCtl(t)
	defer cleanup()

	// Create a check named check1
	check := types.FixtureCheckConfig("check1")
	output, err := sensuctl.run("check", "create", "check1",
		"--command", check.Command,
		"--interval", strconv.FormatUint(uint64(check.Interval), 10),
		"--organization", sensuctl.Organization,
		"--environment", sensuctl.Environment,
		"--subscriptions", strings.Join(check.Subscriptions, ","),
	)
	require.NoError(t, err)
	assert.NotEmpty(t, output)

	// Ensure the check has been created
	output, err = sensuctl.run(
		"check", "info", "check1",
		"--organization", sensuctl.Organization,
		"--environment", sensuctl.Environment,
	)
	require.NoError(t, err)
	c := types.CheckConfig{}
	require.NoError(t, json.Unmarshal(output, &c))
	assert.Equal(t, check.Name, c.Name)

	// Print the list of checks in wrapped-json format
	checkJSON, err := sensuctl.run(
		"check", "list", "--format", "wrapped-json",
		"--organization", sensuctl.Organization,
		"--environment", sensuctl.Environment,
	)
	require.NoError(t, err)
	assert.NotEmpty(t, checkJSON)

	// Delete the check using sensuctl
	_, err = sensuctl.run(
		"check", "delete", "check1", "--skip-confirm",
		"--organization", sensuctl.Organization,
		"--environment", sensuctl.Environment,
	)
	require.NoError(t, err)

	// Ensure the check has been removed
	output, err = sensuctl.run(
		"check", "list",
		"--organization", sensuctl.Organization,
		"--environment", sensuctl.Environment,
	)
	require.NoError(t, err)
	checks := []types.CheckConfig{}
	require.NoError(t, json.Unmarshal(output, &checks))
	assert.Equal(t, 0, len(checks))

	// Use sensuctl create to read the wrapped-json from the file
	sensuctl.stdin = bytes.NewReader(checkJSON)
	_, err = sensuctl.run("create")
	require.NoError(t, err)

	// Ensure the check has been created again
	output, err = sensuctl.run(
		"check", "list", "--format", "json",
		"--organization", sensuctl.Organization,
		"--environment", sensuctl.Environment,
	)
	require.NoError(t, err)
	assert.NotEmpty(t, output)
	checks = []types.CheckConfig{}
	require.NoError(t, json.Unmarshal(output, &checks))
	assert.Equal(t, 1, len(checks))
	assert.Equal(t, "check1", checks[0].Name)
}
