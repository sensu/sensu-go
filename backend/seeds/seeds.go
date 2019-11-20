package seeds

import (
	"context"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/authentication/bcrypt"
	"github.com/sensu/sensu-go/backend/authentication/jwt"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/types"
)

type Config struct {
	// AdminUsername is the username of the cluster admin.
	AdminUsername string

	// AdminPassword is the password of the cluster admin.
	AdminPassword string
}

// SeedClusters seeds the cluster according to the provided config.
func SeedCluster(store store.Store, config Config) error {
	initializer, err := store.NewInitializer()
	if err != nil {
		return err
	}
	logger := logger.WithField("component", "backend.seeds")

	// Lock initialization key to avoid competing installations
	if err = initializer.Lock(); err != nil {
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
	if err := setupAdminUser(store, config.AdminUsername, config.AdminPassword); err != nil {
		logger.WithError(err).Error("could not initialize the admin user")
		return err
	}

	// Create the agent user
	if err := setupAgentUser(store, "agent", "P@ssw0rd!"); err != nil {
		logger.WithError(err).Error("could not initialize the agent user")
		return err
	}

	// Create the default ClusterRoles
	if err := setupClusterRoles(store); err != nil {
		logger.WithError(err).Error("could not initialize the default ClusterRoles and Roles")
		return err
	}

	// Create the default ClusterRoleBindings
	if err := setupClusterRoleBindings(store); err != nil {
		logger.WithError(err).Error("could not initialize the default ClusterRoles and Roles")
		return err
	}

	// Set initialized flag
	return initializer.FlagAsInitialized()
}

// SeedInitialData will seed a store with initial data. This method is
// idempotent and can be safely run every time the backend starts.
func SeedInitialData(store store.Store) (err error) {
	config := Config{
		AdminUsername: "admin",
		AdminPassword: "P@ssw0rd!",
	}
	return SeedCluster(store, config)
}

func setupDefaultNamespace(store store.Store) error {
	return store.CreateNamespace(
		context.Background(),
		&types.Namespace{
			Name: "default",
		})
}

func setupClusterRoleBindings(store store.Store) error {
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
	if err := store.CreateClusterRoleBinding(context.Background(), clusterAdmin); err != nil {
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
	if err := store.CreateClusterRoleBinding(context.Background(), systemAgent); err != nil {
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
	return store.CreateClusterRoleBinding(context.Background(), systemUser)
}

func setupClusterRoles(store store.Store) error {
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
	if err := store.CreateClusterRole(context.Background(), clusterAdmin); err != nil {
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
	if err := store.CreateClusterRole(context.Background(), admin); err != nil {
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
	if err := store.CreateClusterRole(context.Background(), edit); err != nil {
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
	if err := store.CreateClusterRole(context.Background(), view); err != nil {
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
	if err := store.CreateClusterRole(context.Background(), systemAgent); err != nil {
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
	return store.CreateClusterRole(context.Background(), systemUser)
}

func setupAdminUser(store store.Store, username, password string) error {
	hash, err := bcrypt.HashPassword(password)
	if err != nil {
		return err
	}

	admin := &types.User{
		Username: username,
		Password: hash,
		Groups:   []string{"cluster-admins"},
	}
	return store.CreateUser(admin)
}

func setupAgentUser(store store.Store, username, password string) error {
	hash, err := bcrypt.HashPassword("P@ssw0rd!")
	if err != nil {
		return err
	}

	agent := &types.User{
		Username: username,
		Password: hash,
		Groups:   []string{"system:agents"},
	}
	return store.CreateUser(agent)
}
