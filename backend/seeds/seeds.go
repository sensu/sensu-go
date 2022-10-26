package seeds

import (
	"context"
	"errors"

	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

type Config struct {
	// AdminUsername is the username of the cluster admin.
	AdminUsername string

	// AdminPassword is the password of the cluster admin.
	AdminPassword string

	// AdminAPIKey is the API key of the cluster admin. Can be used instead of
	// AdminUsername and AdminPassword.
	AdminAPIKey string
}

var ErrAlreadyInitialized = errors.New("sensu-backend already initialized")

func seedCluster(config Config) storev2.InitializeFunc {
	logger := logger.WithField("component", "backend.seeds")
	logger.Info("seeding store with initial data")

	return func(ctx context.Context, configStore storev2.Interface, nsStore storev2.NamespaceStore) error {
		_, err := nsStore.Get(ctx, "default")
		if err == nil {
			return ErrAlreadyInitialized
		}
		if _, ok := err.(*store.ErrNotFound); !ok {
			return err
		}
		if err := setupNamespaces(ctx, nsStore, config); err != nil {
			return err
		}
		if err := setupUsers(ctx, configStore, config); err != nil {
			return err
		}
		if err := setupAPIKeys(ctx, configStore, config); err != nil {
			return err
		}
		if err := setupClusterRoles(ctx, configStore, config); err != nil {
			return err
		}
		if err := setupClusterRoleBindings(ctx, configStore, config); err != nil {
			return err
		}

		return nil
	}
}

// SeedCluster seeds the cluster according to the provided config.
func SeedCluster(ctx context.Context, s storev2.Interface, nsStore storev2.NamespaceStore, config Config) (fErr error) {
	return s.Initialize(ctx, seedCluster(config))
}

// SeedInitialDataWithContext is like SeedInitialData except it takes an existing
// context.
func SeedInitialDataWithContext(ctx context.Context, s storev2.Interface, nsStore storev2.NamespaceStore) (err error) {
	config := Config{
		AdminUsername: "admin",
		AdminPassword: "P@ssw0rd!",
	}
	return SeedCluster(ctx, s, nsStore, config)
}

func createResource[R storev2.Resource[T], T any](ctx context.Context, s storev2.Interface, resource R) error {
	resourceStore := storev2.NewGenericStore[R](s)
	return resourceStore.CreateIfNotExists(ctx, resource)
}
