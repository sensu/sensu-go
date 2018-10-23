package seeds

import (
	"context"

	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/internal/apis/meta"
	"github.com/sensu/sensu-go/internal/apis/rbac"
	storev2 "github.com/sensu/sensu-go/storage"
	"github.com/sensu/sensu-go/types"
)

// SeedInitialData will seed a store with initial data. This method is
// idempotent and can be safely run every time the backend starts.
func SeedInitialData(store store.Store, storev2 storev2.Store) (err error) {
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

	// Create the default namespace
	if err := setupDefaultNamespace(store); err != nil {
		logger.WithError(err).Error("unable to setup 'default' namespace")
		return err
	}

	// Create the admin user
	if err := setupAdminUser(store); err != nil {
		logger.WithError(err).Error("unable to setup admin user")
		return err
	}

	// Create the sensu read-only user
	if err := setupReadOnlyUser(store); err != nil {
		logger.WithError(err).Error("unable to setup sensu user")
		return err
	}

	// Create the agent user
	if err := setupAgentUser(store); err != nil {
		logger.WithError(err).Error("unable to setup agent user")
		return err
	}

	// Create the admin cluster role
	if err := setupAdminClusterRole(storev2); err != nil {
		logger.WithError(err).Error("unable to setup the cluster-admin role")
		return err
	}

	// Create the read-only role
	if err := setupReadOnlyRole(storev2); err != nil {
		logger.WithError(err).Error("unable to setup the read-only role")
		return err
	}

	// Create the agent role
	if err := setupAgentRole(storev2); err != nil {
		logger.WithError(err).Error("unable to setup the agent role")
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
		Groups:   []string{"cluster-admin"},
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

func setupAgentUser(store store.Store) error {
	// default agent user/pass
	agent := &types.User{
		Username: "agent",
		Password: "P@ssw0rd!",
		Groups:   []string{"agent"},
	}

	return store.CreateUser(agent)
}

func setupAdminClusterRole(store storev2.Store) error {
	role := &rbac.ClusterRole{
		TypeMeta: meta.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: "cluster-admin",
		},
		Rules: []rbac.Rule{
			rbac.Rule{
				APIGroups: []string{"*"},
				Resources: []string{"*"},
				Verbs:     []string{"*"},
			},
		},
	}
	rolebinding := rbac.ClusterRoleBinding{
		TypeMeta: meta.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: "cluster-admin-binding",
		},
		RoleRef: rbac.RoleRef{
			Type: "ClusterRole",
			Name: "cluster-admin",
		},
		Subjects: []rbac.Subject{
			rbac.Subject{
				Kind: rbac.GroupKind,
				Name: "cluster-admin",
			},
		},
	}

	if err := store.Create(context.Background(), "clusterroles/"+role.Name, role); err != nil {
		return err
	}
	if err := store.Create(context.Background(), "clusterrolebindings/"+role.Name, rolebinding); err != nil {
		return err
	}

	return nil
}

func setupReadOnlyRole(store storev2.Store) error {
	role := &rbac.Role{
		TypeMeta: meta.TypeMeta{
			Kind:       "Role",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: "read-only",
		},
		Rules: []rbac.Rule{
			rbac.Rule{
				APIGroups: []string{"core"},
				Resources: []string{"*"},
				Verbs:     []string{"get", "list"},
			},
		},
	}
	rolebinding := rbac.ClusterRoleBinding{
		TypeMeta: meta.TypeMeta{
			Kind:       "RoleBinding",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: "read-only-binding",
		},
		RoleRef: rbac.RoleRef{
			Type: "Role",
			Name: "read-only",
		},
		Subjects: []rbac.Subject{
			rbac.Subject{
				Kind: rbac.GroupKind,
				Name: "read-only",
			},
		},
	}

	if err := store.Create(context.Background(), "roles/"+role.Name, role); err != nil {
		return err
	}
	if err := store.Create(context.Background(), "rolebindings/"+role.Name, rolebinding); err != nil {
		return err
	}

	return nil
}

func setupAgentRole(store storev2.Store) error {
	role := &rbac.Role{
		TypeMeta: meta.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: "agent",
		},
		Rules: []rbac.Rule{
			rbac.Rule{
				APIGroups: []string{"core"},
				Resources: []string{"events"},
				Verbs:     []string{"create", "update"},
			},
		},
	}
	rolebinding := rbac.ClusterRoleBinding{
		TypeMeta: meta.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "v1alpha1",
		},
		ObjectMeta: meta.ObjectMeta{
			Name: "agent-binding",
		},
		RoleRef: rbac.RoleRef{
			Type: "ClusterRole",
			Name: "agent",
		},
		Subjects: []rbac.Subject{
			rbac.Subject{
				Kind: rbac.GroupKind,
				Name: "agent",
			},
		},
	}

	if err := store.Create(context.Background(), "clusterroles/"+role.Name, role); err != nil {
		return err
	}
	if err := store.Create(context.Background(), "clusterrolebindings/"+role.Name, rolebinding); err != nil {
		return err
	}

	return nil
}
