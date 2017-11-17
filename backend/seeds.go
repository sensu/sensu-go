package backend

import (
	"context"

	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// idempotent and can be safely run every time the backend starts.
func seedInitialData(store store.Store) error {
	initializer, _ := store.NewInitializer()
	logger := logger.WithField("component", "backend.seeds")

	// Lock initialization key to avoid competing installations
	if err := initializer.Lock(); err != nil {
		return err
	}
	defer initializer.Close()

	// Initialize the JWT secret. This method is idempotent and needs to be ran
	// at every startup so the JWT signatures remain valid
	if err := jwt.InitSecret(store); err != nil {
		return err
	}

	// Check that the store hasn't already been seeded
	if initialized, err := initializer.IsInitialized(); err != nil {
		return err
	} else if initialized {
		return nil
	}
	logger.Info("seeding etcd store w/ intial data")

	// Set default role
	if err := setupAdminRole(store); err != nil {
		logger.WithError(err).Error("unable to setup admin role")
		return err
	}

	// Default user
	if err := setupDefaultUser(store); err != nil {
		logger.WithError(err).Error("unable to setup admin user")
		return err
	}

	// Default Agent user
	if err := setupDefaultAgentUser(store); err != nil {
		logger.WithError(err).Error("unable to setup agent user")
		return err
	}

	// Default organization
	if err := setupDefaultOrganization(store); err != nil {
		logger.WithError(err).Error("unable to setup 'default' organization")
		return err
	}

	// Default environment
	if err := setupDefaultEnvironment(store); err != nil {
		logger.WithError(err).Error("unable to setup 'default' environment")
		return err
	}

	// Set initialized flag
	return initializer.FlagAsInitialized()
}

func setupAdminRole(store store.Store) error {
	return store.UpdateRole(
		context.Background(),
		&types.Role{
			Name: "admin",
			Rules: []types.Rule{{
				Type:         "*",
				Environment:  "*",
				Organization: "*",
				Permissions:  types.RuleAllPerms,
			}},
		},
	)
}

func setupDefaultEnvironment(store store.Store) error {
	return store.UpdateEnvironment(
		context.Background(),
		"default",
		&types.Environment{
			Name:        "default",
			Description: "Default environment",
		})
}

func setupDefaultOrganization(store store.Store) error {
	return store.UpdateOrganization(
		context.Background(),
		&types.Organization{
			Name:        "default",
			Description: "Default organization",
		})
}

func setupDefaultUser(store store.Store) error {
	// Set default user
	admin := &types.User{
		Username: "admin",
		Password: "P@ssw0rd!",
		Roles:    []string{"admin"},
	}

	return store.CreateUser(admin)
}

func setupDefaultAgentUser(store store.Store) error {
	// default agent user/pass
	agent := &types.User{
		Username: "agent",
		Password: "P@ssw0rd!",
		Roles:    []string{"agent"},
	}

	return store.CreateUser(agent)
}
