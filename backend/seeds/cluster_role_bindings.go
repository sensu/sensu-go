package seeds

import (
	"context"
	"errors"
	"fmt"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
)

func setupClusterRoleBindings(ctx context.Context, s storev2.Interface, config Config) error {
	clusterRoleBindings := []*corev2.ClusterRoleBinding{
		clusterAdminClusterRoleBinding(),
		systemAgentClusterRoleBinding(),
		systemUserClusterRoleBinding(),
	}

	for _, clusterRoleBinding := range clusterRoleBindings {
		name := clusterRoleBinding.ObjectMeta.Name

		if err := createResource(ctx, s, clusterRoleBinding); err != nil {
			var alreadyExists *store.ErrAlreadyExists
			if !errors.As(err, &alreadyExists) {
				msg := fmt.Sprintf("could not initialize the %s cluster role binding", name)
				logger.WithError(err).Error(msg)
				return fmt.Errorf("%s: %w", msg, err)
			}
			logger.Warnf("%s cluster role binding already exists", name)
		}
	}

	return nil
}

func clusterAdminClusterRoleBinding() *corev2.ClusterRoleBinding {
	// The cluster-admin ClusterRoleBinding grants permission found in the
	// cluster-admin ClusterRole to any user belonging to the cluster-admins group
	return &corev2.ClusterRoleBinding{
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
}

func systemAgentClusterRoleBinding() *corev2.ClusterRoleBinding {
	// The system:agent ClusterRoleBinding grants permission found in the
	// system-agent ClusterRole to any agents belonging to the system:agents group
	return &corev2.ClusterRoleBinding{
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
}

func systemUserClusterRoleBinding() *corev2.ClusterRoleBinding {
	// The system:user ClusterRoleBinding grants permission found in the
	// system:user ClusterRole to any user belonging to the system:users group
	return &corev2.ClusterRoleBinding{
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
}
