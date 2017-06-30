package backend

import (
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// SeedInitialData seeds initial data into the store. Ideally this operation is
// idempotent and can be safely run every time the backend starts.
func seedInitialData(store store.Store) error {
	// Default role
	if err := setupDefaultRole(store); err != nil {
		return err
	}

	return nil
}

func setupDefaultRole(store store.Store) error {
	return store.UpdateRole(&types.Role{
		Name: "admin",
		Rules: []types.Rule{{
			Type:         "*",
			Organization: "*",
			Permissions:  types.RuleAllPerms,
		}},
	})
}
