package schema

//go:generate go run github.com/sensu/sensu-go/scripts/gen_gqltype -types AssetBuild,ClusterRole,ClusterRoleBinding,Pipeline,PipelineWorkflow,ResourceReference,Role,RoleBinding,RoleRef,Rule,Secret,Subject -pkg-path ../../../../api/core/v2 -o ./corev2.gen.graphql
//go:generate go run github.com/sensu/sensu-go/scripts/gengraphql .
