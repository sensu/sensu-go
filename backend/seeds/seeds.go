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

	// Set admin role
	// if err := setupAdminRole(store); err != nil {
	// 	logger.WithError(err).Error("unable to setup admin role")
	// 	return err
	// }

	// Set read-only role
	// if err := setupReadOnlyRole(store); err != nil {
	// 	logger.WithError(err).Error("unable to setup read-only role")
	// 	return err
	// }

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

	// Default admin cluster role & binding
	if err := setupAdminClusterRole(storev2); err != nil {
		logger.WithError(err).Error("unable to setup the cluster admin role")
		return err
	}

	// Default read-only role & binding
	if err := setupReadOnlyRole(storev2); err != nil {
		logger.WithError(err).Error("unable to setup the read-only role")
		return err
	}

	// Default namespace
	if err := setupDefaultNamespace(store); err != nil {
		logger.WithError(err).Error("unable to setup 'default' namespace")
		return err
	}

	// Set initialized flag
	return initializer.FlagAsInitialized()
}

// func setupAdminRole(store store.Store) error {
// 	return store.UpdateRole(
// 		context.Background(),
// 		&types.Role{
// 			Name: "admin",
// 			Rules: []types.Rule{{
// 				Type:        types.RuleTypeAll,
// 				Namespace:   types.NamespaceTypeAll,
// 				Permissions: types.RuleAllPerms,
// 			}},
// 		},
// 	)
// }

// func setupReadOnlyRole(store store.Store) error {
// 	return store.UpdateRole(
// 		context.Background(),
// 		&types.Role{
// 			Name: "read-only",
// 			Rules: []types.Rule{{
// 				Type:        types.RuleTypeAll,
// 				Namespace:   types.NamespaceTypeAll,
// 				Permissions: []string{types.RulePermRead},
// 			}},
// 		},
// 	)
// }

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
		Roles:    []string{"admin"},
		Groups:   []string{},
	}

	return store.CreateUser(admin)
}

func setupReadOnlyUser(store store.Store) error {
	// Set default read-only user
	sensu := &types.User{
		Username: "sensu",
		Password: "sensu",
		Roles:    []string{"read-only"},
		Groups:   []string{},
	}

	return store.CreateUser(sensu)
}

func setupDefaultAgentUser(store store.Store) error {
	// default agent user/pass
	agent := &types.User{
		Username: "agent",
		Password: "P@ssw0rd!",
		Roles:    []string{"agent"},
		Groups:   []string{},
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
				Kind: rbac.UserKind,
				Name: "admin",
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
			Kind:       "rRoleBinding",
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
				Kind: rbac.UserKind,
				Name: "sensu",
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
