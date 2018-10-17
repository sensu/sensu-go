package e2e

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newRBACSensuctl(t *testing.T, wsURL, httpURL, namespace, user, pass string) (*sensuCtl, func()) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "sensuctl")
	if err != nil {
		t.Fatal(err)
	}

	ctl := &sensuCtl{
		Namespace: namespace,
		ConfigDir: tmpDir,
		stdin:     os.Stdin,
		wsURL:     wsURL,
		httpURL:   httpURL,
	}

	// Authenticate sensuctl
	out, err := ctl.run("configure",
		"-n",
		"--url", httpURL,
		"--username", user,
		"--password", pass,
		"--format", "json",
		"--namespace", "default",
	)
	if err != nil {
		t.Fatal(err, string(out))
	}

	return ctl, func() { _ = os.RemoveAll(tmpDir) }
}

func TestRBAC(t *testing.T) {
	t.Skip("skip")
	t.Parallel()

	// Start the backend
	backend, cleanup, err := newBackendProcess(40010, 40011, 40012, 40013, 40014)
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	require.NoError(t, backend.Start())

	if !waitForBackend(backend.HTTPURL) {
		t.Fatal("backend not ready")
	}

	// Initializes sensuctl
	adminctl, cleanup := newCustomSensuctl(t, backend.WSURL, backend.HTTPURL, "default")
	defer cleanup()

	// Make sure we are properly authenticated
	output, err := adminctl.run("user", "list")
	assert.NoError(t, err)

	users := []types.User{}
	require.NoError(t, json.Unmarshal(output, &users))
	assert.NotZero(t, len(users))

	// Create the following hierarchy for RBAC:
	// -- default (namespace)
	//        -- default-check (check)
	//        -- default-handler (handler)
	// -- acme (namespace)
	//        -- dev-check (check)
	//        -- dev-handler (handler)
	output, err = adminctl.run("namespace", "create", "acme",
		"--description", "acme",
	)
	assert.NoError(t, err, string(output))

	defaultCheck := types.FixtureCheckConfig("default-check")
	output, err = adminctl.run("check", "create", defaultCheck.Name,
		"--command", defaultCheck.Command,
		"--interval", strconv.FormatUint(uint64(defaultCheck.Interval), 10),
		"--runtime-assets", strings.Join(defaultCheck.RuntimeAssets, ","),
		"--subscriptions", strings.Join(defaultCheck.Subscriptions, ","),
		"--namespace", defaultCheck.Namespace,
		"--publish",
	)
	assert.NoError(t, err, string(output))

	devCheck := types.FixtureCheckConfig("dev-check")
	devCheck.Namespace = "acme"
	output, err = adminctl.run("check", "create", devCheck.Name,
		"--command", devCheck.Command,
		"--interval", strconv.FormatUint(uint64(devCheck.Interval), 10),
		"--runtime-assets", strings.Join(devCheck.RuntimeAssets, ","),
		"--subscriptions", strings.Join(devCheck.Subscriptions, ","),
		"--namespace", devCheck.Namespace,
		"--publish",
	)
	assert.NoError(t, err, string(output))

	checkHook := types.FixtureHookList("hook1")
	output, err = adminctl.run("check", "set-hooks", defaultCheck.Name,
		"--namespace", defaultCheck.Namespace,
		"--type", checkHook.Type,
		"--hooks", strings.Join(checkHook.Hooks, ","),
	)
	assert.NoError(t, err, string(output))

	output, err = adminctl.run("check", "set-hooks", devCheck.Name,
		"--namespace", devCheck.Namespace,
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
		"--namespace", defaultHandler.Namespace,
	)
	assert.NoError(t, err, string(output))

	devHandler := types.FixtureHandler("dev-handler")
	devHandler.Namespace = "acme"
	output, err = adminctl.run("handler", "create", devHandler.Name,
		"--type", devHandler.Type,
		"--mutator", devHandler.Mutator,
		"--command", devHandler.Command,
		"--timeout", strconv.FormatUint(uint64(devHandler.Timeout), 10),
		"--socket-host", "",
		"--socket-port", "",
		"--handlers", strings.Join(devHandler.Handlers, ","),
		"--namespace", devHandler.Namespace,
	)
	assert.NoError(t, err, string(output))

	// Create roles for every namespace
	defaultRole := types.FixtureRole("default", "default")
	output, err = adminctl.run("role", "create", defaultRole.Name,
		"--type", defaultRole.Rules[0].Type,
		"-crud",
	)
	assert.NoError(t, err, string(output))

	devRole := types.FixtureRole("dev", "acme")
	output, err = adminctl.run("role", "create", devRole.Name,
		"--type", devRole.Rules[0].Type,
		"--namespace", devRole.Rules[0].Namespace,
		"-crud",
	)
	assert.NoError(t, err, string(output))

	// Create users for every namespace
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

	// Create a Sensu client for every namespace
	defaultctl, cleanup := newRBACSensuctl(t, backend.WSURL, backend.HTTPURL, "default", "default", "P@ssw0rd!")
	defer cleanup()

	devctl, cleanup := newRBACSensuctl(t, backend.WSURL, backend.HTTPURL, "acme", "dev", "P@ssw0rd!")
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

	// A user with all privileges should be able to query all checks
	checks = []types.CheckConfig{}
	output, err = adminctl.run("check", "list", "--all-namespaces")
	assert.NoError(t, err, string(output))
	require.NoError(t, json.Unmarshal(output, &checks))
	assert.Len(t, checks, 3)

	// A user with all privileges should be able to query a specific namespace
	checks = []types.CheckConfig{}
	output, err = adminctl.run("check", "list", "--namespace", "acme")
	assert.NoError(t, err, string(output))
	require.NoError(t, json.Unmarshal(output, &checks))
	assert.Len(t, checks, 2)

	// A user with all privileges should be able to query a specific namespace
	checks = []types.CheckConfig{}
	output, err = adminctl.run("check", "list", "--namespace", "acme")
	assert.NoError(t, err, string(output))
	require.NoError(t, json.Unmarshal(output, &checks))
	assert.Len(t, checks, 1)

	// Make sure a client can't create objects outside of its role
	output, err = devctl.run("check", "create", defaultCheck.Name,
		"--command", defaultCheck.Command,
		"--interval", strconv.FormatUint(uint64(defaultCheck.Interval), 10),
		"--runtime-assets", strings.Join(defaultCheck.RuntimeAssets, ","),
		"--subscriptions", strings.Join(defaultCheck.Subscriptions, ","),
		"--namespace", defaultCheck.Namespace,
	)
	assert.Error(t, err, string(output))

	// Make sure a client can delete objects within its role
	output, err = devctl.run("check", "delete", devCheck.Name,
		"--namespace", devCheck.Namespace,
		"--skip-confirm",
	)
	assert.NoError(t, err, string(output))

	// Now we want to restart the backend to make sure the JWT will continue
	// to work and prevent an issue like https://github.com/sensu/sensu-go/issues/502
	require.NoError(t, backend.Terminate())
	err = backend.Start()
	if err != nil {
		log.Panic(err)
	}

	// Make sure the backend is available
	backendIsOnline := waitForBackend(backend.HTTPURL)
	assert.True(t, backendIsOnline)

	// Make sure we are properly authenticated
	output, err = adminctl.run("user", "list")
	assert.NoError(t, err, string(output))

}
