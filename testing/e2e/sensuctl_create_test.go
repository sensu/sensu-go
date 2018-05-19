package e2e

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestSensuctlCreate(t *testing.T) {
	t.Parallel()

	// Start the backend
	backend, cleanup := newBackend(t)
	defer cleanup()

	// Initializes sensuctl
	sensuctl, cleanup := newSensuCtl(backend.HTTPURL, "default", "default", "admin", "P@ssw0rd!")
	defer cleanup()

	// Create a check named check1
	check := types.FixtureCheckConfig("check1")
	output, err := sensuctl.run("check", "create", "check1",
		"--command", check.Command,
		"--interval", strconv.FormatUint(uint64(check.Interval), 10),
		"--organization", check.Organization,
		"--environment", check.Environment,
		"--subscriptions", strings.Join(check.Subscriptions, ","),
	)
	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// Ensure the check has been created
	output, err = sensuctl.run("check", "info", "check1")
	assert.NoError(t, err)
	c := types.CheckConfig{}
	assert.NoError(t, json.Unmarshal(output, &c))
	assert.Equal(t, check.Name, c.Name)

	// Print the list of checks in wrapped-json format
	output, err = sensuctl.run("check", "list", "--format", "wrapped-json")
	assert.NoError(t, err)
	assert.NotEmpty(t, output)

	// Write the wrapped-json list to a temp file called checks.json
	file, cleanup := writeTempFile(t, output, "checks.json")
	defer cleanup()

	// Delete the check using sensuctl
	_, err = sensuctl.run("check", "delete", "check1", "--skip-confirm")
	assert.NoError(t, err)

	// Ensure the check has been removed
	output, err = sensuctl.run("check", "list")
	assert.NoError(t, err)
	checks := []types.CheckConfig{}
	assert.NoError(t, json.Unmarshal(output, &checks))
	assert.Equal(t, 0, len(checks))

	// Use sensuctl create to read the wrapped-json from the file
	_, err = sensuctl.run("create", "-f", file)
	assert.NoError(t, err)

	// Ensure the check has been created again
	output, err = sensuctl.run("check", "list", "--format", "json")
	assert.NoError(t, err)
	assert.NotEmpty(t, output)
	checks = []types.CheckConfig{}
	assert.NoError(t, json.Unmarshal(output, &checks))
	assert.Equal(t, 1, len(checks))
	assert.Equal(t, "check1", checks[0].Name)
}
