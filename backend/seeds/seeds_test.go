package seeds

import (
	"context"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	corev3 "github.com/sensu/sensu-go/api/core/v3"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/backend/store/v2/storetest"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSeedInitialDataWithContext(t *testing.T) {
	ctx := context.Background()

	config := Config{
		AdminUsername: "admin",
		AdminPassword: "P@ssw0rd!",
	}

	// Setup store
	s := new(storetest.Store)
	s.On("CreateIfNotExists", mock.Anything, mock.Anything).Return(nil)
	s.On("Initialize", mock.Anything, mock.Anything).Return(seedCluster(ctx, s, config)(ctx))

	sErr := SeedInitialDataWithContext(ctx, s)
	require.NoError(t, sErr, "seeding process should not raise an error")

	err := SeedInitialDataWithContext(ctx, s)
	if err != ErrAlreadyInitialized {
		require.NoError(t, err, "seeding process should be able to be run more than once without error")
	}

	userStoreName := (&corev2.User{}).StoreName()
	namespaceStoreName := (&corev3.Namespace{}).StoreName()

	// ensure the default namespace is created
	s.AssertCalled(t, "CreateIfNotExists",
		storev2.NewResourceRequest(ctx, "", "default", namespaceStoreName),
		mock.Anything)

	// ensure the admin user is created
	s.AssertCalled(t, "CreateIfNotExists",
		storev2.NewResourceRequest(ctx, "", "admin", userStoreName),
		mock.Anything)

	// ensure the agent user is created
	s.AssertCalled(t, "CreateIfNotExists",
		storev2.NewResourceRequest(ctx, "", "agent", userStoreName),
		mock.Anything)

	// ensure the cluster-admin cluster role is created
	clusterAdminClusterRole := &corev2.ClusterRole{
		ObjectMeta: corev2.NewObjectMeta("cluster-admin", ""),
		Rules: []corev2.Rule{
			{
				Verbs:     []string{corev2.VerbAll},
				Resources: []string{corev2.ResourceAll},
			},
		},
	}
	s.AssertCalled(t, "CreateIfNotExists",
		storev2.NewResourceRequestFromResource(ctx, clusterAdminClusterRole),
		mock.Anything)

	// ensure the admin cluster role is created
	adminClusterRole := &corev2.ClusterRole{
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
	s.AssertCalled(t, "CreateIfNotExists",
		storev2.NewResourceRequestFromResource(ctx, adminClusterRole),
		mock.Anything)

	// ensure the edit cluster role is created
	editClusterRole := &corev2.ClusterRole{
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
	s.AssertCalled(t, "CreateIfNotExists",
		storev2.NewResourceRequestFromResource(ctx, editClusterRole),
		mock.Anything)

	// ensure the view cluster role is created
	viewClusterRole := &corev2.ClusterRole{
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
	s.AssertCalled(t, "CreateIfNotExists",
		storev2.NewResourceRequestFromResource(ctx, viewClusterRole),
		mock.Anything)

	// ensure the system:agent cluster role is created
	systemAgentClusterRole := &corev2.ClusterRole{
		ObjectMeta: corev2.NewObjectMeta("system:agent", ""),
		Rules: []corev2.Rule{
			{
				Verbs:     []string{corev2.VerbAll},
				Resources: []string{"events"},
			},
		},
	}
	s.AssertCalled(t, "CreateIfNotExists",
		storev2.NewResourceRequestFromResource(ctx, systemAgentClusterRole),
		mock.Anything)

	// ensure the system:user cluster role is created
	systemUserClusterRole := &corev2.ClusterRole{
		ObjectMeta: corev2.NewObjectMeta("system:user", ""),
		Rules: []corev2.Rule{
			{
				Verbs:     []string{"get", "update"},
				Resources: []string{corev2.LocalSelfUserResource},
			},
		},
	}
	s.AssertCalled(t, "CreateIfNotExists",
		storev2.NewResourceRequestFromResource(ctx, systemUserClusterRole),
		mock.Anything)

	// ensure the cluster-admin cluster role binding is created
	clusterAdminClusterRoleBinding := &corev2.ClusterRoleBinding{
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
	s.AssertCalled(t, "CreateIfNotExists",
		storev2.NewResourceRequestFromResource(ctx, clusterAdminClusterRoleBinding),
		mock.Anything)

	// ensure the system:agent cluster role binding is created
	systemAgentClusterRoleBinding := &corev2.ClusterRoleBinding{
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
	s.AssertCalled(t, "CreateIfNotExists",
		storev2.NewResourceRequestFromResource(ctx, systemAgentClusterRoleBinding),
		mock.Anything)

	// ensure the system:user cluster role binding is created
	systemUserClusterRoleBinding := &corev2.ClusterRoleBinding{
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
	s.AssertCalled(t, "CreateIfNotExists",
		storev2.NewResourceRequestFromResource(ctx, systemUserClusterRoleBinding),
		mock.Anything)
}
