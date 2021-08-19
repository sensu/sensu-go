package seeds

import (
	"context"
	"errors"
	"fmt"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authentication/bcrypt"
	storev1 "github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/backend/store/etcd"
	"github.com/sensu/sensu-go/types"
	clientv3 "go.etcd.io/etcd/client/v3"
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

// SeedCluster seeds the cluster according to the provided config.
func SeedCluster(ctx context.Context, store storev1.Store, client *clientv3.Client, config Config) (fErr error) {
	logger := logger.WithField("component", "backend.seeds")

	initializer, err := store.NewInitializer(ctx)
	if err != nil {
		return fmt.Errorf("failed to create seed initializer: %w", err)
	}

	// Lock initialization key to avoid competing installations
	if err = initializer.Lock(ctx); err != nil {
		return fmt.Errorf("failed to create initializer lock: %w", err)
	}
	defer func() {
		if err := initializer.Close(ctx); fErr == nil && err != nil {
			fErr = fmt.Errorf("failed to close initializer: %w", err)
		}
	}()

	// Check that the store hasn't already been initialized
	initialized, err := initializer.IsInitialized(ctx)
	if err != nil {
		return fmt.Errorf("failed to check if cluster has been initialized: %w", err)
	}

	if initialized {
		logger.Info("store already initialized")
		return ErrAlreadyInitialized
	}

	logger.Info("seeding etcd store with initial data")

	// Create the default namespace
	if err := setupDefaultNamespace(ctx, store); err != nil {
		switch err := err.(type) {
		case *storev1.ErrAlreadyExists:
			logger.Warn("default namespace already exists")
		default:
			msg := "unable to setup default namespace"
			logger.WithError(err).Error(msg)
			return fmt.Errorf("%s: %w", msg, err)
		}
	}

	// Create the admin user
	if err := setupAdminUser(ctx, store, config.AdminUsername, config.AdminPassword, config.AdminAPIKey); err != nil {
		switch err := err.(type) {
		case *storev1.ErrAlreadyExists:
			logger.Warn("admin user already exists")
		default:
			msg := "could not initialize the admin user"
			logger.WithError(err).Error(msg)
			return fmt.Errorf("%s: %w", msg, err)
		}
	}

	// Create the agent user
	if err := setupAgentUser(ctx, store, "agent", "P@ssw0rd!"); err != nil {
		switch err := err.(type) {
		case *storev1.ErrAlreadyExists:
			logger.Warn("agent user already exists")
		default:
			msg := "could not initialize the agent user"
			logger.WithError(err).Error(msg)
			return fmt.Errorf("%s: %w", msg, err)
		}
	}

	// Create the default ClusterRoles
	if err := setupClusterRoles(ctx, store); err != nil {
		switch err := err.(type) {
		case *storev1.ErrAlreadyExists:
			logger.Warn("default ClusterRoles and Roles already exist")
		default:
			msg := "could not initialize the default ClusterRoles and Roles"
			logger.WithError(err).Error(msg)
			return fmt.Errorf("%s: %w", msg, err)
		}
	}

	// Create the default ClusterRoleBindings
	if err := setupClusterRoleBindings(ctx, store); err != nil {
		switch err := err.(type) {
		case *storev1.ErrAlreadyExists:
			logger.Warn("default ClusterRoleBindings already exist")
		default:
			msg := "could not initialize the default ClusterRoleBindings"
			logger.WithError(err).Error(msg)
			return fmt.Errorf("%s: %w", msg, err)
		}
	}

	if client != nil {
		// Migrate the cluster to the latest version
		if err := etcd.MigrateDB(ctx, client, etcd.Migrations); err != nil {
			logger.WithError(err).Error("error bringing the database to the latest version")
			return fmt.Errorf("error bringing the database to the latest version: %w", err)
		}
		if len(etcd.EnterpriseMigrations) > 0 {
			if err = etcd.MigrateEnterpriseDB(ctx, client, etcd.EnterpriseMigrations); err != nil {
				logger.WithError(err).Error("error bringing the enterprise database to the latest version")
				return
			}
		}
	}

	// Set initialized flag
	return initializer.FlagAsInitialized(ctx)
}

// SeedInitialData will seed a store with initial data. This method is
// idempotent and can be safely run every time the backend starts.
func SeedInitialData(store storev1.Store) (err error) {
	return SeedInitialDataWithContext(context.Background(), store)
}

// SeedInitialDataWithContext is like SeedInitialData except it takes an existing
// context.
func SeedInitialDataWithContext(ctx context.Context, store storev1.Store) (err error) {
	config := Config{
		AdminUsername: "admin",
		AdminPassword: "P@ssw0rd!",
	}
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	return SeedCluster(ctx, store, nil, config)
}

func setupDefaultNamespace(ctx context.Context, store storev1.Store) error {
	return store.CreateNamespace(
		ctx,
		&types.Namespace{
			Name: "default",
		})
}

func setupClusterRoleBindings(ctx context.Context, store storev1.Store) error {
	// The cluster-admin ClusterRoleBinding grants permission found in the
	// cluster-admin ClusterRole to any user belonging to the cluster-admins group
	clusterAdmin := &types.ClusterRoleBinding{
		ObjectMeta: corev2.NewObjectMeta("cluster-admin", ""),
		RoleRef: types.RoleRef{
			Type: "ClusterRole",
			Name: "cluster-admin",
		},
		Subjects: []types.Subject{
			types.Subject{
				Type: "Group",
				Name: "cluster-admins",
			},
		},
	}
	if err := store.CreateClusterRoleBinding(ctx, clusterAdmin); err != nil {
		return err
	}

	// The system:agent ClusterRoleBinding grants permission found in the
	// system-agent ClusterRole to any agents belonging to the system:agents group
	systemAgent := &types.ClusterRoleBinding{
		ObjectMeta: corev2.NewObjectMeta("system:agent", ""),
		RoleRef: types.RoleRef{
			Type: "ClusterRole",
			Name: "system:agent",
		},
		Subjects: []types.Subject{
			types.Subject{
				Type: "Group",
				Name: "system:agents",
			},
		},
	}
	if err := store.CreateClusterRoleBinding(ctx, systemAgent); err != nil {
		return err
	}

	// The system:user ClusterRoleBinding grants permission found in the
	// system:user ClusterRole to any user belonging to the system:users group
	systemUser := &types.ClusterRoleBinding{
		ObjectMeta: corev2.NewObjectMeta("system:user", ""),
		RoleRef: types.RoleRef{
			Type: "ClusterRole",
			Name: "system:user",
		},
		Subjects: []types.Subject{
			types.Subject{
				Type: "Group",
				Name: "system:users",
			},
		},
	}
	return store.CreateClusterRoleBinding(ctx, systemUser)
}

func setupClusterRoles(ctx context.Context, store storev1.Store) error {
	// The cluster-admin ClusterRole gives access to perform any action on any
	// resource. When used in a ClusterRoleBinding, it gives full control over
	// every resource in the cluster and in all namespaces. When used in a
	// RoleBinding, it gives full control over every resource in the rolebinding's
	// namespace, including the namespace itself
	clusterAdmin := &types.ClusterRole{
		ObjectMeta: corev2.NewObjectMeta("cluster-admin", ""),
		Rules: []types.Rule{
			types.Rule{
				Verbs:     []string{types.VerbAll},
				Resources: []string{types.ResourceAll},
			},
		},
	}
	if err := store.CreateClusterRole(ctx, clusterAdmin); err != nil {
		return err
	}

	// The admin ClusterRole is intended to be used within a namespace using a
	// RoleBinding. It gives full access to most resources, including the ability
	// to create Roles and RoleBindings within the namespace but does not allow
	// write access to the namespace itself
	admin := &types.ClusterRole{
		ObjectMeta: corev2.NewObjectMeta("admin", ""),
		Rules: []types.Rule{
			types.Rule{
				Verbs: []string{types.VerbAll},
				Resources: append(types.CommonCoreResources, []string{
					"roles",
					"rolebindings",
				}...),
			},
			types.Rule{
				Verbs: []string{"get", "list"},
				Resources: []string{
					"namespaces",
				},
			},
		},
	}
	if err := store.CreateClusterRole(ctx, admin); err != nil {
		return err
	}

	// The edit ClusterRole is intended to be used within a namespace using a
	// RoleBinding. It allows read/write access to most objects in a namespace. It
	// does not allow viewing or modifying roles or rolebindings.
	edit := &types.ClusterRole{
		ObjectMeta: corev2.NewObjectMeta("edit", ""),
		Rules: []types.Rule{
			types.Rule{
				Verbs:     []string{types.VerbAll},
				Resources: types.CommonCoreResources,
			},
			types.Rule{
				Verbs: []string{"get", "list"},
				Resources: []string{
					"namespaces",
				},
			},
		},
	}
	if err := store.CreateClusterRole(ctx, edit); err != nil {
		return err
	}

	// The view ClusterRole is intended to be used within a namespace using a
	// RoleBinding. It allows read-only access to see most objects in a namespace.
	// It does not allow viewing roles or rolebindings.
	view := &types.ClusterRole{
		ObjectMeta: corev2.NewObjectMeta("view", ""),
		Rules: []types.Rule{
			types.Rule{
				Verbs: []string{"get", "list"},
				Resources: append(types.CommonCoreResources, []string{
					"namespaces",
				}...),
			},
		},
	}
	if err := store.CreateClusterRole(ctx, view); err != nil {
		return err
	}

	// The systemAgent ClusterRole is used by Sensu agents and should not be
	// modified by the users. Modification to this ClusterRole can result in
	// non-functional Sensu agents.
	systemAgent := &types.ClusterRole{
		ObjectMeta: corev2.NewObjectMeta("system:agent", ""),
		Rules: []types.Rule{
			types.Rule{
				Verbs:     []string{types.VerbAll},
				Resources: []string{"events"},
			},
		},
	}
	if err := store.CreateClusterRole(ctx, systemAgent); err != nil {
		return err
	}

	// The systemUser ClusterRole is used by local users and should not be
	// modified by the users. Modification to his ClusterRole can result in
	// non-functional Sensu users. It allows users to view themselves and change
	// their own password
	systemUser := &types.ClusterRole{
		ObjectMeta: corev2.NewObjectMeta("system:user", ""),
		Rules: []types.Rule{
			types.Rule{
				Verbs:     []string{"get", "update"},
				Resources: []string{types.LocalSelfUserResource},
			},
		},
	}
	return store.CreateClusterRole(ctx, systemUser)
}

func setupAdminUser(ctx context.Context, store storev1.Store, username, password, apiKey string) error {
	hash, err := bcrypt.HashPassword(password)
	if err != nil {
		return err
	}

	admin := &types.User{
		Username:     username,
		Password:     hash,
		PasswordHash: hash,
		Groups:       []string{"cluster-admins"},
	}
	if err := store.CreateUser(ctx, admin); err != nil {
		return err
	}
	if apiKey != "" {
		key := &corev2.APIKey{
			ObjectMeta: corev2.ObjectMeta{
				Name:      apiKey,
				CreatedBy: username,
			},
			Username:  username,
			CreatedAt: time.Now().Unix(),
		}
		if err := store.CreateResource(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

func setupAgentUser(ctx context.Context, store storev1.Store, username, password string) error {
	hash, err := bcrypt.HashPassword("P@ssw0rd!")
	if err != nil {
		return err
	}

	agent := &types.User{
		Username:     username,
		Password:     hash,
		PasswordHash: hash,
		Groups:       []string{"system:agents"},
	}
	return store.CreateUser(ctx, agent)
}
