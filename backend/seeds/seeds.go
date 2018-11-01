package seeds

import (
	"context"

	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// SeedInitialData will seed a store with initial data. This method is
// idempotent and can be safely run every time the backend starts.
func SeedInitialData(store store.Store) (err error) {
	initializer, _ := store.NewInitializer()
	logger := logger.WithField("component", "backend.seeds")

	// Lock initialization key to avoid competing installations
	if err := initializer.Lock(); err != nil {
		return err
	}
	defer func() {
		e := initializer.Close()
		if err == nil {
			err = e
		}
	}()

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

	// Admin user
	if err := setupAdminUser(store); err != nil {
		logger.WithError(err).Error("unable to setup admin user")
		return err
	}

	// Default read-only user (sensu)
	if err := setupReadOnlyUser(store); err != nil {
		logger.WithError(err).Error("unable to setup sensu user")
		return err
	}

	// Default Agent user
	if err := setupDefaultAgentUser(store); err != nil {
		logger.WithError(err).Error("unable to setup agent user")
		return err
	}

	// Default namespace & environment
	if err := setupDefaultNamespace(store); err != nil {
		logger.WithError(err).Error("unable to setup 'default' namespace")
		return err
	}

	// Set initialized flag
	return initializer.FlagAsInitialized()
}

func setupDefaultNamespace(store store.Store) error {
	return store.CreateNamespace(
		context.Background(),
		&types.Namespace{
			Name: "default",
		})
}

func setupAdminUser(store store.Store) error {
	// Setup admin user
	admin := &types.User{
		Username: "admin",
		Password: "P@ssw0rd!",
		Groups:   []string{"admin"},
	}

	return store.CreateUser(admin)
}

func setupReadOnlyUser(store store.Store) error {
	// Set default read-only user
	sensu := &types.User{
		Username: "sensu",
		Password: "sensu",
		Groups:   []string{"read-only"},
	}

	return store.CreateUser(sensu)
}

func setupDefaultAgentUser(store store.Store) error {
	// default agent user/pass
	agent := &types.User{
		Username: "agent",
		Password: "P@ssw0rd!",
		Groups:   []string{"agent"},
	}

	return store.CreateUser(agent)
}
