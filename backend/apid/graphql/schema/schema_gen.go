package schema

//go:generate go run github.com/sensu/sensu-go/scripts/gen_gqltype -types AssetBuild,ClusterRole,ClusterRoleBinding,Deregistration,Network,NetworkInterface,Pipeline,PipelineWorkflow,Process,ResourceReference,Role,RoleBinding,RoleRef,Rule,Secret,Subject,System -pkg-path github.com/sensu/core/v2 -o ./corev2.gen.graphql
//go:generate go run github.com/sensu/sensu-go/scripts/gen_gqltype -types EntityConfig,EntityState -pkg-path github.com/sensu/core/v3 -o ./corev3.gen.graphql
//go:generate go run github.com/sensu/sensu-go/scripts/gengraphql .
