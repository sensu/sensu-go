package e2e

import (
	"encoding/json"
	"log"
	"strconv"
	"strings"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRBAC(t *testing.T) {
	t.Parallel()

	// Start the backend
	backend, cleanup := newBackend(t)
	defer cleanup()

	// Initializes sensuctl as admin
	adminctl, cleanup := newSensuCtl(backend.HTTPURL, "default", "default", "admin", "P@ssw0rd!")
	defer cleanup()

	// Make sure we are properly authenticated
	output, err := adminctl.run("user", "list")
	assert.NoError(t, err)

	users := []types.User{}
	require.NoError(t, json.Unmarshal(output, &users))
	assert.NotZero(t, len(users))

	// Create the following hierarchy for RBAC:
	// -- default (organization)
	//    -- default (environment)
	//        -- default-check (check)
	//        -- default-handler (handler)
	// -- acme (organization)
	//    -- dev (environment)
	//        -- dev-check (check)
	//        -- dev-handler (handler)
	//    -- prod (environment)
	//        -- prod-check (check)
	//        -- prod-handler (handler)

	output, err = adminctl.run("organization", "create", "acme",
		"--description", "acme",
	)
	assert.NoError(t, err, string(output))

	output, err = adminctl.run("environment", "create", "default",
		"--organization", "acme",
	)
	assert.NoError(t, err, string(output))

	output, err = adminctl.run("environment", "create", "dev",
		"--organization", "acme",
	)
	assert.NoError(t, err, string(output))

	output, err = adminctl.run("environment", "create", "prod",
		"--organization", "acme",
	)
	assert.NoError(t, err, string(output))

	defaultCheck := types.FixtureCheckConfig("default-check")
	output, err = adminctl.run("check", "create", defaultCheck.Name,
		"--command", defaultCheck.Command,
		"--interval", strconv.FormatUint(uint64(defaultCheck.Interval), 10),
		"--runtime-assets", strings.Join(defaultCheck.RuntimeAssets, ","),
		"--subscriptions", strings.Join(defaultCheck.Subscriptions, ","),
		"--organization", defaultCheck.Organization,
		"--environment", defaultCheck.Environment,
		"--publish",
	)
	assert.NoError(t, err, string(output))

	devCheck := types.FixtureCheckConfig("dev-check")
	devCheck.Organization = "acme"
	devCheck.Environment = "dev"
	output, err = adminctl.run("check", "create", devCheck.Name,
		"--command", devCheck.Command,
		"--interval", strconv.FormatUint(uint64(devCheck.Interval), 10),
		"--runtime-assets", strings.Join(devCheck.RuntimeAssets, ","),
		"--subscriptions", strings.Join(devCheck.Subscriptions, ","),
		"--organization", devCheck.Organization,
		"--environment", devCheck.Environment,
		"--publish",
	)
	assert.NoError(t, err, string(output))

	prodCheck := types.FixtureCheckConfig("prod-check")
	prodCheck.Organization = "acme"
	prodCheck.Environment = "prod"
	output, err = adminctl.run("check", "create", prodCheck.Name,
		"--command", prodCheck.Command,
		"--interval", strconv.FormatUint(uint64(prodCheck.Interval), 10),
		"--runtime-assets", strings.Join(prodCheck.RuntimeAssets, ","),
		"--subscriptions", strings.Join(prodCheck.Subscriptions, ","),
		"--organization", prodCheck.Organization,
		"--environment", prodCheck.Environment,
		"--publish",
	)
	assert.NoError(t, err, string(output))

	checkHook := types.FixtureHookList("hook1")
	output, err = adminctl.run("check", "set-hooks", defaultCheck.Name,
		"--organization", defaultCheck.Organization,
		"--environment", defaultCheck.Environment,
		"--type", checkHook.Type,
		"--hooks", strings.Join(checkHook.Hooks, ","),
	)
	assert.NoError(t, err, string(output))

	output, err = adminctl.run("check", "set-hooks", devCheck.Name,
		"--organization", devCheck.Organization,
		"--environment", devCheck.Environment,
		"--type", checkHook.Type,
		"--hooks", strings.Join(checkHook.Hooks, ","),
	)
	assert.NoError(t, err, string(output))

	output, err = adminctl.run("check", "set-hooks", prodCheck.Name,
		"--organization", prodCheck.Organization,
		"--environment", prodCheck.Environment,
		"--type", checkHook.Type,
		"--hooks", strings.Join(checkHook.Hooks, ","),
	)
	assert.NoError(t, err, string(output))

	defaultHandler := types.FixtureHandler("default-handler")
	output, err = adminctl.run("handler", "create", defaultHandler.Name,
		"--type", defaultHandler.Type,
		"--mutator", defaultHandler.Mutator,
		"--command", defaultHandler.Command,
		"--timeout", strconv.FormatUint(uint64(defaultHandler.Timeout), 10),
		"--socket-host", "",
		"--socket-port", "",
		"--handlers", strings.Join(defaultHandler.Handlers, ","),
		"--organization", defaultHandler.Organization,
		"--environment", defaultHandler.Environment,
	)
	assert.NoError(t, err, string(output))

	devHandler := types.FixtureHandler("dev-handler")
	devHandler.Organization = "acme"
	devHandler.Environment = "dev"
	output, err = adminctl.run("handler", "create", devHandler.Name,
		"--type", devHandler.Type,
		"--mutator", devHandler.Mutator,
		"--command", devHandler.Command,
		"--timeout", strconv.FormatUint(uint64(devHandler.Timeout), 10),
		"--socket-host", "",
		"--socket-port", "",
		"--handlers", strings.Join(devHandler.Handlers, ","),
		"--organization", devHandler.Organization,
		"--environment", devHandler.Environment,
	)
	assert.NoError(t, err, string(output))

	prodHandler := types.FixtureHandler("prod-handler")
	prodHandler.Organization = "acme"
	prodHandler.Environment = "prod"
	_, err = adminctl.run("handler", "create", prodHandler.Name,
		"--type", prodHandler.Type,
		"--mutator", prodHandler.Mutator,
		"--command", prodHandler.Command,
		"--timeout", strconv.FormatUint(uint64(prodHandler.Timeout), 10),
		"--socket-host", "",
		"--socket-port", "",
		"--handlers", strings.Join(prodHandler.Handlers, ","),
		"--organization", prodHandler.Organization,
		"--environment", prodHandler.Environment,
	)
	assert.NoError(t, err, string(output))

	// Create roles for every environment
	defaultRole := types.FixtureRole("default", "default", "default")
	output, err = adminctl.run("role", "create", defaultRole.Name)
	assert.NoError(t, err, string(output))
	output, err = adminctl.run("role", "add-rule", defaultRole.Name,
		"--type", defaultRole.Rules[0].Type,
		"--organization", defaultRole.Rules[0].Organization,
		"--environment", defaultRole.Rules[0].Environment,
		"-crud",
	)
	assert.NoError(t, err, string(output))

	devRole := types.FixtureRole("dev", "acme", "dev")
	output, err = adminctl.run("role", "create", devRole.Name)
	assert.NoError(t, err, string(output))
	output, err = adminctl.run("role", "add-rule", devRole.Name,
		"--type", devRole.Rules[0].Type,
		"--organization", devRole.Rules[0].Organization,
		"--environment", devRole.Rules[0].Environment,
		"-crud",
	)
	assert.NoError(t, err, string(output))

	prodRole := types.FixtureRole("prod", "acme", "prod")
	output, err = adminctl.run("role", "create", prodRole.Name)
	assert.NoError(t, err, string(output))
	output, err = adminctl.run("role", "add-rule", prodRole.Name,
		"--type", prodRole.Rules[0].Type,
		"--organization", prodRole.Rules[0].Organization,
		"--environment", prodRole.Rules[0].Environment,
		"-crud",
	)
	assert.NoError(t, err, string(output))

	// Create users for every environment
	defaultUser := types.FixtureUser("default")
	defaultUser.Roles = []string{defaultRole.Name}
	output, err = adminctl.run("user", "create", defaultUser.Username,
		"--password", defaultUser.Password,
		"--roles", strings.Join(defaultUser.Roles, ","),
	)
	assert.NoError(t, err, string(output))

	devUser := types.FixtureUser("dev")
	devUser.Roles = []string{devRole.Name}
	output, err = adminctl.run("user", "create", devUser.Username,
		"--password", devUser.Password,
		"--roles", strings.Join(devUser.Roles, ","),
	)
	assert.NoError(t, err, string(output))

	prodUser := types.FixtureUser("prod")
	prodUser.Roles = []string{prodRole.Name}
	output, err = adminctl.run("user", "create", prodUser.Username,
		"--password", prodUser.Password,
		"--roles", strings.Join(prodUser.Roles, ","),
	)
	assert.NoError(t, err, string(output))

	// Create a Sensu client for every environment
	defaultctl, cleanup := newSensuCtl(backend.HTTPURL, "default", "default", "default", "P@ssw0rd!")
	defer cleanup()

	devctl, cleanup := newSensuCtl(backend.HTTPURL, "acme", "dev", "dev", "P@ssw0rd!")
	defer cleanup()

	prodctl, cleanup := newSensuCtl(backend.HTTPURL, "acme", "prod", "prod", "P@ssw0rd!")
	defer cleanup()

	// Make sure each of these clients only has access to objects within its role
	checks := []types.CheckConfig{}
	output, err = defaultctl.run("check", "list")
	assert.NoError(t, err, string(output))
	require.NoError(t, json.Unmarshal(output, &checks))
	assert.Equal(t, defaultCheck, &checks[0])

	checks = []types.CheckConfig{}
	output, err = devctl.run("check", "list")
	assert.NoError(t, err, string(output))
	require.NoError(t, json.Unmarshal(output, &checks))
	assert.Equal(t, devCheck, &checks[0])

	checks = []types.CheckConfig{}
	output, err = prodctl.run("check", "list")
	assert.NoError(t, err, string(output))
	require.NoError(t, json.Unmarshal(output, &checks))
	assert.Equal(t, prodCheck, &checks[0])

	// A user with all privileges should be able to query all checks
	checks = []types.CheckConfig{}
	output, err = adminctl.run("check", "list", "--environment", "*", "--all-organizations")
	// output, err = adminctl.run("check", "list", "--organization", "*", "--environment", "*")
	assert.NoError(t, err, string(output))
	require.NoError(t, json.Unmarshal(output, &checks))
	assert.Len(t, checks, 3)

	// A user with all privileges should be able to query a specific organization
	checks = []types.CheckConfig{}
	output, err = adminctl.run("check", "list", "--environment", "*", "--organization", "acme")
	assert.NoError(t, err, string(output))
	require.NoError(t, json.Unmarshal(output, &checks))
	assert.Len(t, checks, 2)

	// A user with all privileges should be able to query a specific organization
	// and a specific environment
	checks = []types.CheckConfig{}
	output, err = adminctl.run("check", "list", "--environment", "dev", "--organization", "acme")
	assert.NoError(t, err, string(output))
	require.NoError(t, json.Unmarshal(output, &checks))
	assert.Len(t, checks, 1)

	// Make sure a client can't create objects outside of its role
	output, err = devctl.run("check", "create", defaultCheck.Name,
		"--command", defaultCheck.Command,
		"--interval", strconv.FormatUint(uint64(defaultCheck.Interval), 10),
		"--runtime-assets", strings.Join(defaultCheck.RuntimeAssets, ","),
		"--subscriptions", strings.Join(defaultCheck.Subscriptions, ","),
		"--organization", defaultCheck.Organization,
		"--environment", defaultCheck.Environment,
	)
	assert.Error(t, err, string(output))

	output, err = devctl.run("check", "create", prodCheck.Name,
		"--command", prodCheck.Command,
		"--interval", strconv.FormatUint(uint64(prodCheck.Interval), 10),
		"--runtime-assets", strings.Join(prodCheck.RuntimeAssets, ","),
		"--subscriptions", strings.Join(prodCheck.Subscriptions, ","),
		"--organization", prodCheck.Organization,
		"--environment", prodCheck.Environment,
	)
	assert.Error(t, err, string(output))

	// Make sure a client can delete objects within its role
	output, err = devctl.run("check", "delete", devCheck.Name,
		"--organization", devCheck.Organization,
		"--environment", devCheck.Environment,
		"--skip-confirm",
	)
	assert.NoError(t, err, string(output))

	// Make sure a client can't delete objects outside of its role
	output, err = devctl.run("check", "delete", prodCheck.Name,
		"--organization", prodCheck.Organization,
		"--environment", prodCheck.Environment,
		"--skip-confirm",
	)
	assert.Error(t, err, string(output))

	// Make sure a client can't read objects outside of its role
	_, err = devctl.run("check", "info", prodCheck.Name)
	assert.Error(t, err)

	// Now we want to restart the backend to make sure the JWT will continue
	// to work and prevent an issue like https://github.com/sensu/sensu-go/issues/502
	require.NoError(t, backend.Kill())
	err = backend.Start()
	if err != nil {
		log.Panic(err)
	}

	// Make sure the backend is available
	backendIsOnline := waitForBackend(backend.HTTPURL)
	assert.True(t, backendIsOnline)

	// Make sure we are properly authenticated
	output, err = adminctl.run("user", "list")
	assert.NoError(t, err, output)

}
