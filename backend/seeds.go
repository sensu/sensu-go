package backend

import (
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

// SeedInitialData seeds initial data into the store. Ideally this operation is
// idempotent and can be safely run every time the backend starts.
func seedInitialData(store store.Store) error {
	if err := setDefaultRole(store); err != nil {
		return err
	}

	return nil
}

func setDefaultRole(store store.Store) error {
	return store.CreateRole(&types.Role{
		Name: "default",
		Rules: []types.Rule{{
			Type:         "*",
			Organization: "*",
			Permissions:  types.RuleAllPerms,
		}},
	})
}
