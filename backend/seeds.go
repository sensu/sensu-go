package backend

import (
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// SeedInitialData seeds initial data into the store. Ideally this operation is
// idempotent and can be safely run every time the backend starts.
func seedInitialData(store store.Store) error {
	initializer, _ := store.NewInitializer()

	// Lock initialization key to avoid competing installations
	if err := initializer.Lock(); err != nil {
		return err
	}
	defer initializer.Close()

	// Check that the store hasn't already been seeded
	if initialized, err := initializer.IsInitialized(); err != nil {
		return err
	} else if initialized {
		return nil
	}

	logger.Debug("seeding etcd store w/ intial data")

	// Set default role
	if err := setupAdminRole(store); err != nil {
		return err
	}

	// Default organization
	if err := setupDefaultOrganization(store); err != nil {
		return err
	}

	// Set initialized flag
	if err := initializer.FlagAsInitialized(); err != nil {
		return err
	}

	return nil
}

func setupAdminRole(store store.Store) error {
	return store.UpdateRole(&types.Role{
		Name: "admin",
		Rules: []types.Rule{{
			Type:         "*",
			Organization: "*",
			Permissions:  types.RuleAllPerms,
		}},
	})
}

func setupDefaultOrganization(store store.Store) error {
	return store.UpdateOrganization(&types.Organization{
		Name:        "default",
		Description: "Default organization",
	})
}
