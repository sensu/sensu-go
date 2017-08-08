package e2e

import (
	"fmt"
	"log"
	"testing"

	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/client/config/basic"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestRBAC(t *testing.T) {
	// Start the backend
	bep, cleanup := newBackendProcess()
	defer cleanup()

	err := bep.Start()
	if err != nil {
		log.Panic(err)
	}
	defer bep.Kill()

	// Make sure the backend is available
	backendHTTPURL := fmt.Sprintf("http://127.0.0.1:%d", bep.APIPort)
	backendIsOnline := waitForBackend(backendHTTPURL)
	assert.True(t, backendIsOnline)

	// Create an authenticated client using the admin user
	adminConfig := &basic.Config{
		Cluster: basic.Cluster{
			APIUrl: backendHTTPURL,
		},
	}
	adminClient := client.New(adminConfig)
	adminTokens, _ := adminClient.CreateAccessToken(backendHTTPURL, "admin", "P@ssw0rd!")
	adminConfig.Cluster.Tokens = adminTokens

	// Make sure we are properly authenticated
	users, err := adminClient.ListUsers()
	assert.NoError(t, err)
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
	org := &types.Organization{Name: "acme"}
	if err := adminClient.CreateOrganization(org); err != nil {
		assert.Fail(t, err.Error())
	}
	env := &types.Environment{Name: "dev"}
	if err := adminClient.CreateEnvironment("acme", env); err != nil {
		assert.Fail(t, err.Error())
	}
	env = &types.Environment{Name: "prod"}
	if err := adminClient.CreateEnvironment("acme", env); err != nil {
		assert.Fail(t, err.Error())
	}

	defaultCheck := types.FixtureCheckConfig("default-check")
	if err := adminClient.CreateCheck(defaultCheck); err != nil {
		assert.Fail(t, err.Error())
	}

	devCheck := types.FixtureCheckConfig("dev-check")
	devCheck.Organization = "acme"
	devCheck.Environment = "dev"
	if err := adminClient.CreateCheck(devCheck); err != nil {
		assert.Fail(t, err.Error())
	}

	prodCheck := types.FixtureCheckConfig("prod-check")
	prodCheck.Organization = "acme"
	prodCheck.Environment = "prod"
	if err := adminClient.CreateCheck(prodCheck); err != nil {
		assert.Fail(t, err.Error())
	}

	defaultHandler := types.FixtureHandler("default-handler")
	if err := adminClient.CreateHandler(defaultHandler); err != nil {
		assert.Fail(t, err.Error())
	}

	devHandler := types.FixtureHandler("dev-handler")
	devHandler.Organization = "acme"
	devHandler.Environment = "dev"
	if err := adminClient.CreateHandler(devHandler); err != nil {
		assert.Fail(t, err.Error())
	}

	prodHandler := types.FixtureHandler("prod-handler")
	prodHandler.Organization = "acme"
	prodHandler.Environment = "prod"
	if err := adminClient.CreateHandler(prodHandler); err != nil {
		assert.Fail(t, err.Error())
	}

	// Create roles for every environment
	defaultRole := types.FixtureRole("default", "default", "default")
	if err := adminClient.CreateRole(defaultRole); err != nil {
		assert.Fail(t, err.Error())
	}
	devRole := types.FixtureRole("dev", "acme", "dev")
	if err := adminClient.CreateRole(devRole); err != nil {
		assert.Fail(t, err.Error())
	}
	prodRole := types.FixtureRole("prod", "acme", "prod")
	if err := adminClient.CreateRole(prodRole); err != nil {
		assert.Fail(t, err.Error())
	}

	// Create users for every environment
	defaultUser := types.FixtureUser("default")
	defaultUser.Roles = []string{defaultRole.Name}
	if err := adminClient.CreateUser(defaultUser); err != nil {
		assert.Fail(t, err.Error())
	}
	devUser := types.FixtureUser("dev")
	devUser.Roles = []string{devRole.Name}
	if err := adminClient.CreateUser(devUser); err != nil {
		assert.Fail(t, err.Error())
	}
	prodUser := types.FixtureUser("prod")
	prodUser.Roles = []string{prodRole.Name}
	if err := adminClient.CreateUser(prodUser); err != nil {
		assert.Fail(t, err.Error())
	}

	// Create a Sensu client for every environment
	// TODO: Simplify the client creation!
	defaultConfig := &basic.Config{
		Cluster: basic.Cluster{
			APIUrl: backendHTTPURL,
		},
	}
	defaultClient := client.New(defaultConfig)
	defaultTokens, _ := defaultClient.CreateAccessToken(backendHTTPURL, "default", "P@ssw0rd!")
	defaultConfig.Cluster.Tokens = defaultTokens

	devConfig := &basic.Config{
		Cluster: basic.Cluster{
			APIUrl: backendHTTPURL,
		},
		Profile: basic.Profile{
			Environment:  "dev",
			Organization: "acme",
		},
	}
	devClient := client.New(devConfig)
	devTokens, _ := devClient.CreateAccessToken(backendHTTPURL, "dev", "P@ssw0rd!")
	devConfig.Cluster.Tokens = devTokens

	prodConfig := &basic.Config{
		Cluster: basic.Cluster{
			APIUrl: backendHTTPURL,
		},
		Profile: basic.Profile{
			Environment:  "prod",
			Organization: "acme",
		},
	}
	prodClient := client.New(prodConfig)
	prodTokens, _ := prodClient.CreateAccessToken(backendHTTPURL, "prod", "P@ssw0rd!")
	prodConfig.Cluster.Tokens = prodTokens

	// Make sure each of these clients only has access to objects within its role
	checks, err := defaultClient.ListChecks()
	assert.NoError(t, err)
	assert.Equal(t, &checks[0], defaultCheck)

	checks, err = devClient.ListChecks()
	assert.NoError(t, err)
	assert.Equal(t, &checks[0], devCheck)

	checks, err = prodClient.ListChecks()
	assert.NoError(t, err)
	assert.Equal(t, &checks[0], prodCheck)

	// Make sure a client can't create objects outside of its role
	if err := devClient.CreateCheck(defaultCheck); err == nil {
		assert.Fail(t, "devClient should not be able to create into the default org")
	}

	if err := devClient.CreateCheck(prodCheck); err == nil {
		assert.Fail(t, "devClient should not be able to create into the prod env")
	}

	// Make sure a client can't delete objects outside of its role
	if err := devClient.DeleteCheck(prodCheck); err == nil {
		assert.Fail(t, "devClient should not be able to delete into the prod env")
	}

	// Make sure a client can't read objects outside of its role
	// TODO (Simon): We should be able to override the env without saving it
	devConfig.SaveEnvironment("prod")
	if _, err := devClient.FetchCheck(prodCheck.Name); err == nil {
		assert.Fail(t, "devClient should not be able to read into the prod env")
	}
	devConfig.SaveEnvironment("dev")
}
