package seeds

import (
	"context"
	"testing"

	corev2 "github.com/sensu/core/v2"
	corev3 "github.com/sensu/core/v3"
	"github.com/sensu/sensu-go/backend/store"
	storev2 "github.com/sensu/sensu-go/backend/store/v2"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestSeedInitialDataWithContext(t *testing.T) {
	ctx := context.Background()

	config := Config{
		AdminUsername: "admin",
		AdminPassword: "P@ssw0rd!",
	}

	// Setup stores
	nsStore := new(mockstore.NamespaceStore)
	nsStore.On("CreateIfNotExists", mock.Anything, mock.Anything).Return(nil)
	nsStore.On("Get", mock.Anything, "default").Return((*corev3.Namespace)(nil), new(store.ErrNotFound))

	s := new(mockstore.V2MockStore)
	cs := new(mockstore.ConfigStore)
	s.On("GetNamespaceStore").Return(nsStore)
	s.On("GetConfigStore").Return(cs)
	cs.On("CreateIfNotExists", mock.Anything, mock.Anything, mock.Anything).Return(nil)
	cs.On("Initialize", mock.Anything, mock.Anything).Return(seedCluster(config)(ctx, s))

	sErr := SeedInitialDataWithContext(ctx, s)
	require.NoError(t, sErr, "seeding process should not raise an error")

	err := SeedInitialDataWithContext(ctx, s)
	if err != ErrAlreadyInitialized {
		require.NoError(t, err, "seeding process should be able to be run more than once without error")
	}

	// store names
	userStoreName := (&corev2.User{}).StoreName()

	// type metas
	userTypeMeta := corev2.TypeMeta{Type: "User", APIVersion: "core/v2"}

	// ensure the default namespace is created
	nsStore.AssertCalled(t, "CreateIfNotExists", context.Background(), defaultNamespace())

	// ensure the admin user is created
	cs.AssertCalled(t, "CreateIfNotExists",
		context.Background(),
		storev2.NewResourceRequest(userTypeMeta, "", "admin", userStoreName),
		mock.Anything)

	// ensure the agent user is created
	cs.AssertCalled(t, "CreateIfNotExists",
		context.Background(),
		storev2.NewResourceRequest(userTypeMeta, "", "agent", userStoreName),
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
	cs.AssertCalled(t, "CreateIfNotExists",
		context.Background(),
		storev2.NewResourceRequestFromResource(clusterAdminClusterRole),
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
	cs.AssertCalled(t, "CreateIfNotExists",
		context.Background(),
		storev2.NewResourceRequestFromResource(adminClusterRole),
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
	cs.AssertCalled(t, "CreateIfNotExists",
		context.Background(),
		storev2.NewResourceRequestFromResource(editClusterRole),
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
	cs.AssertCalled(t, "CreateIfNotExists",
		context.Background(),
		storev2.NewResourceRequestFromResource(viewClusterRole),
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
	cs.AssertCalled(t, "CreateIfNotExists",
		context.Background(),
		storev2.NewResourceRequestFromResource(systemAgentClusterRole),
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
	cs.AssertCalled(t, "CreateIfNotExists",
		context.Background(),
		storev2.NewResourceRequestFromResource(systemUserClusterRole),
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
	cs.AssertCalled(t, "CreateIfNotExists",
		context.Background(),
		storev2.NewResourceRequestFromResource(clusterAdminClusterRoleBinding),
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
	cs.AssertCalled(t, "CreateIfNotExists",
		context.Background(),
		storev2.NewResourceRequestFromResource(systemAgentClusterRoleBinding),
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
	cs.AssertCalled(t, "CreateIfNotExists",
		context.Background(),
		storev2.NewResourceRequestFromResource(systemUserClusterRoleBinding),
		mock.Anything)
}
