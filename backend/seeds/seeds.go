package seeds

import (
	"context"
	"errors"
	"fmt"

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

func seedCluster(ctx context.Context, s storev2.Interface, config Config) func(context.Context) error {
	logger := logger.WithField("component", "backend.seeds")
	logger.Info("seeding etcd store with initial data")

	return func(context.Context) error {
		if err := setupNamespaces(ctx, s, config); err != nil {
			return err
		}
		if err := setupUsers(ctx, s, config); err != nil {
			return err
		}
		if err := setupAPIKeys(ctx, s, config); err != nil {
			return err
		}
		if err := setupClusterRoles(ctx, s, config); err != nil {
			return err
		}
		if err := setupClusterRoleBindings(ctx, s, config); err != nil {
			return err
		}

		if err := setupClusterRoleBindings(ctx, s, config); err != nil {
			var alreadyExists *store.ErrAlreadyExists
			if !errors.As(err, &alreadyExists) {
				msg := "could not initialize the default ClusterRoleBindings"
				logger.WithError(err).Error(msg)
				return fmt.Errorf("%s: %w", msg, err)
			}
			logger.Warn("default ClusterRoleBindings already exist")
		}
		return nil
	}
}

// SeedCluster seeds the cluster according to the provided config.
func SeedCluster(ctx context.Context, s storev2.Interface, config Config) (fErr error) {
	return s.Initialize(ctx, seedCluster(ctx, s, config))
}

// SeedInitialDataWithContext is like SeedInitialData except it takes an existing
// context.
func SeedInitialDataWithContext(ctx context.Context, s storev2.Interface) (err error) {
	config := Config{
		AdminUsername: "admin",
		AdminPassword: "P@ssw0rd!",
	}
	return SeedCluster(ctx, s, config)
}

func createResource[R storev2.Resource[T], T any](ctx context.Context, s storev2.Interface, resource R) error {
	resourceStore := storev2.NewGenericStore[R](s)
	return resourceStore.CreateIfNotExists(ctx, resource)
}
