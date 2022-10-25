package seeds

import (
	"context"
	"errors"
	"fmt"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

func setupClusterRoles(ctx context.Context, s storev2.Interface, config Config) error {
	clusterRoles := []*corev2.ClusterRole{
		clusterAdminClusterRole(),
		adminClusterRole(),
		editClusterRole(),
		viewClusterRole(),
		systemAgentClusterRole(),
		systemUserClusterRole(),
	}

	for _, clusterRole := range clusterRoles {
		name := clusterRole.ObjectMeta.Name

		if err := createResource(ctx, s, clusterRole); err != nil {
			var alreadyExists *store.ErrAlreadyExists
			if !errors.As(err, &alreadyExists) {
				msg := fmt.Sprintf("could not initialize the %s cluster role", name)
				logger.WithError(err).Error(msg)
				return fmt.Errorf("%s: %w", msg, err)
			}
			logger.Warnf("%s cluster role already exists", name)
		}
	}

	return nil
}

func clusterAdminClusterRole() *corev2.ClusterRole {
	// The cluster-admin ClusterRole gives access to perform any action on any
	// resource. When used in a ClusterRoleBinding, it gives full control over
	// every resource in the cluster and in all namespaces. When used in a
	// RoleBinding, it gives full control over every resource in the rolebinding's
	// namespace, including the namespace itself
	return &corev2.ClusterRole{
		ObjectMeta: corev2.NewObjectMeta("cluster-admin", ""),
		Rules: []corev2.Rule{
			{
				Verbs:     []string{corev2.VerbAll},
				Resources: []string{corev2.ResourceAll},
			},
		},
	}
}

func adminClusterRole() *corev2.ClusterRole {
	// The admin ClusterRole is intended to be used within a namespace using a
	// RoleBinding. It gives full access to most resources, including the ability
	// to create Roles and RoleBindings within the namespace but does not allow
	// write access to the namespace itself
	return &corev2.ClusterRole{
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
}

func editClusterRole() *corev2.ClusterRole {
	// The edit ClusterRole is intended to be used within a namespace using a
	// RoleBinding. It allows read/write access to most objects in a namespace. It
	// does not allow viewing or modifying roles or rolebindings.
	return &corev2.ClusterRole{
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
}

func viewClusterRole() *corev2.ClusterRole {
	// The view ClusterRole is intended to be used within a namespace using a
	// RoleBinding. It allows read-only access to see most objects in a namespace.
	// It does not allow viewing roles or rolebindings.
	return &corev2.ClusterRole{
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
}

func systemAgentClusterRole() *corev2.ClusterRole {
	// The systemAgent ClusterRole is used by Sensu agents and should not be
	// modified by the users. Modification to this ClusterRole can result in
	// non-functional Sensu agents.
	return &corev2.ClusterRole{
		ObjectMeta: corev2.NewObjectMeta("system:agent", ""),
		Rules: []corev2.Rule{
			{
				Verbs:     []string{corev2.VerbAll},
				Resources: []string{"events"},
			},
		},
	}
}

func systemUserClusterRole() *corev2.ClusterRole {
	// The systemUser ClusterRole is used by local users and should not be
	// modified by the users. Modification to his ClusterRole can result in
	// non-functional Sensu users. It allows users to view themselves and change
	// their own password
	return &corev2.ClusterRole{
		ObjectMeta: corev2.NewObjectMeta("system:user", ""),
		Rules: []corev2.Rule{
			{
				Verbs:     []string{"get", "update"},
				Resources: []string{corev2.LocalSelfUserResource},
			},
		},
	}
}
