package seeds

import (
	"context"
	"errors"
	"fmt"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	"github.com/sensu/sensu-go/backend/authentication/bcrypt"
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
		// Create the default namespace
		if err := setupDefaultNamespace(ctx, s); err != nil {
			var alreadyExists *store.ErrAlreadyExists
			if !errors.As(err, &alreadyExists) {
				msg := "unable to setup default namespace"
				logger.WithError(err).Error(msg)
				return fmt.Errorf("%s: %w", msg, err)
			}
			logger.Warn("default namespace already exists")
		}

		// Create the admin user
		if err := setupAdminUser(ctx, s, config.AdminUsername, config.AdminPassword, config.AdminAPIKey); err != nil {
			var alreadyExists *store.ErrAlreadyExists
			if !errors.As(err, &alreadyExists) {
				msg := "could not initialize the admin user"
				logger.WithError(err).Error(msg)
				return fmt.Errorf("%s: %w", msg, err)
			}
			logger.Warn("admin user already exists")
		}

		// Create the agent user
		if err := setupAgentUser(ctx, s, "agent", "P@ssw0rd!"); err != nil {
			var alreadyExists *store.ErrAlreadyExists
			if !errors.As(err, &alreadyExists) {
				msg := "could not initialize the agent user"
				logger.WithError(err).Error(msg)
				return fmt.Errorf("%s: %w", msg, err)
			}
			logger.Warn("agent user already exists")
		}

		// Create the default ClusterRoles
		if err := setupClusterRoles(ctx, s); err != nil {
			var alreadyExists *store.ErrAlreadyExists
			if !errors.As(err, &alreadyExists) {
				msg := "could not initialize the default ClusterRoles and Roles"
				logger.WithError(err).Error(msg)
				return fmt.Errorf("%s: %w", msg, err)
			}
			logger.Warn("default ClusterRoles and Roles already exist")
		}

		// Create the default ClusterRoleBindings
		if err := setupClusterRoleBindings(ctx, s); err != nil {
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

func createResources(ctx context.Context, s storev2.Interface, resources []corev3.Resource) error {
	for _, resource := range resources {
		req := storev2.NewResourceRequestFromResource(ctx, resource)
		wrapper, err := storev2.WrapResource(resource)
		if err != nil {
			return err
		}
		if err := s.CreateIfNotExists(req, wrapper); err != nil {
			return err
		}
	}
	return nil
}

func setupDefaultNamespace(ctx context.Context, s storev2.Interface) error {
	namespace := &corev3.Namespace{
		Metadata: &corev2.ObjectMeta{
			Name:        "default",
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		},
	}
	req := storev2.NewResourceRequestFromResource(ctx, namespace)
	wrapper, err := storev2.WrapResource(namespace)
	if err != nil {
		return err
	}
	return s.CreateIfNotExists(req, wrapper)
}

func setupClusterRoleBindings(ctx context.Context, s storev2.Interface) error {
	// The cluster-admin ClusterRoleBinding grants permission found in the
	// cluster-admin ClusterRole to any user belonging to the cluster-admins group
	clusterAdmin := &corev2.ClusterRoleBinding{
		ObjectMeta: corev2.NewObjectMeta("cluster-admin", ""),
		RoleRef: corev2.RoleRef{
			Type: "ClusterRole",
			Name: "cluster-admin",
		},
		Subjects: []corev2.Subject{
			{
				Type: "Group",
				Name: "cluster-admins",
			},
		},
	}

	// The system:agent ClusterRoleBinding grants permission found in the
	// system-agent ClusterRole to any agents belonging to the system:agents group
	systemAgent := &corev2.ClusterRoleBinding{
		ObjectMeta: corev2.NewObjectMeta("system:agent", ""),
		RoleRef: corev2.RoleRef{
			Type: "ClusterRole",
			Name: "system:agent",
		},
		Subjects: []corev2.Subject{
			{
				Type: "Group",
				Name: "system:agents",
			},
		},
	}

	// The system:user ClusterRoleBinding grants permission found in the
	// system:user ClusterRole to any user belonging to the system:users group
	systemUser := &corev2.ClusterRoleBinding{
		ObjectMeta: corev2.NewObjectMeta("system:user", ""),
		RoleRef: corev2.RoleRef{
			Type: "ClusterRole",
			Name: "system:user",
		},
		Subjects: []corev2.Subject{
			{
				Type: "Group",
				Name: "system:users",
			},
		},
	}

	resources := []corev3.Resource{
		clusterAdmin,
		systemAgent,
		systemUser,
	}
	return createResources(ctx, s, resources)
}

func setupClusterRoles(ctx context.Context, s storev2.Interface) error {
	// The cluster-admin ClusterRole gives access to perform any action on any
	// resource. When used in a ClusterRoleBinding, it gives full control over
	// every resource in the cluster and in all namespaces. When used in a
	// RoleBinding, it gives full control over every resource in the rolebinding's
	// namespace, including the namespace itself
	clusterAdmin := &corev2.ClusterRole{
		ObjectMeta: corev2.NewObjectMeta("cluster-admin", ""),
		Rules: []corev2.Rule{
			{
				Verbs:     []string{corev2.VerbAll},
				Resources: []string{corev2.ResourceAll},
			},
		},
	}

	// The admin ClusterRole is intended to be used within a namespace using a
	// RoleBinding. It gives full access to most resources, including the ability
	// to create Roles and RoleBindings within the namespace but does not allow
	// write access to the namespace itself
	admin := &corev2.ClusterRole{
		ObjectMeta: corev2.NewObjectMeta("admin", ""),
		Rules: []corev2.Rule{
			{
				Verbs: []string{corev2.VerbAll},
				Resources: append(corev2.CommonCoreResources, []string{
					"roles",
					"rolebindings",
				}...),
			},
			{
				Verbs: []string{"get", "list"},
				Resources: []string{
					"namespaces",
				},
			},
		},
	}

	// The edit ClusterRole is intended to be used within a namespace using a
	// RoleBinding. It allows read/write access to most objects in a namespace. It
	// does not allow viewing or modifying roles or rolebindings.
	edit := &corev2.ClusterRole{
		ObjectMeta: corev2.NewObjectMeta("edit", ""),
		Rules: []corev2.Rule{
			{
				Verbs:     []string{corev2.VerbAll},
				Resources: corev2.CommonCoreResources,
			},
			{
				Verbs: []string{"get", "list"},
				Resources: []string{
					"namespaces",
				},
			},
		},
	}

	// The view ClusterRole is intended to be used within a namespace using a
	// RoleBinding. It allows read-only access to see most objects in a namespace.
	// It does not allow viewing roles or rolebindings.
	view := &corev2.ClusterRole{
		ObjectMeta: corev2.NewObjectMeta("view", ""),
		Rules: []corev2.Rule{
			{
				Verbs: []string{"get", "list"},
				Resources: append(corev2.CommonCoreResources, []string{
					"namespaces",
				}...),
			},
		},
	}

	// The systemAgent ClusterRole is used by Sensu agents and should not be
	// modified by the users. Modification to this ClusterRole can result in
	// non-functional Sensu agents.
	systemAgent := &corev2.ClusterRole{
		ObjectMeta: corev2.NewObjectMeta("system:agent", ""),
		Rules: []corev2.Rule{
			{
				Verbs:     []string{corev2.VerbAll},
				Resources: []string{"events"},
			},
		},
	}

	// The systemUser ClusterRole is used by local users and should not be
	// modified by the users. Modification to his ClusterRole can result in
	// non-functional Sensu users. It allows users to view themselves and change
	// their own password
	systemUser := &corev2.ClusterRole{
		ObjectMeta: corev2.NewObjectMeta("system:user", ""),
		Rules: []corev2.Rule{
			{
				Verbs:     []string{"get", "update"},
				Resources: []string{corev2.LocalSelfUserResource},
			},
		},
	}

	resources := []corev3.Resource{
		clusterAdmin,
		admin,
		edit,
		view,
		systemAgent,
		systemUser,
	}

	return createResources(ctx, s, resources)
}

func setupAdminUser(ctx context.Context, s storev2.Interface, username, password, apiKey string) error {
	hash, err := bcrypt.HashPassword(password)
	if err != nil {
		return err
	}

	resources := []corev3.Resource{}

	admin := &corev2.User{
		Username:     username,
		Password:     hash,
		PasswordHash: hash,
		Groups:       []string{"cluster-admins"},
	}
	resources = append(resources, admin)

	if apiKey != "" {
		key := &corev2.APIKey{
			ObjectMeta: corev2.ObjectMeta{
				Name:      apiKey,
				CreatedBy: username,
			},
			Username:  username,
			CreatedAt: time.Now().Unix(),
		}
		resources = append(resources, key)
	}

	return createResources(ctx, s, resources)
}

func setupAgentUser(ctx context.Context, s storev2.Interface, username, password string) error {
	hash, err := bcrypt.HashPassword("P@ssw0rd!")
	if err != nil {
		return err
	}

	agent := &corev2.User{
		Username:     username,
		Password:     hash,
		PasswordHash: hash,
		Groups:       []string{"system:agents"},
	}

	return createResources(ctx, s, []corev3.Resource{
		agent,
	})
}
