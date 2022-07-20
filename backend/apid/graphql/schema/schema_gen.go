package schema

//go:generate go run github.com/sensu/sensu-go/scripts/gen_gqltype -types AssetBuild,ClusterRole,ClusterRoleBinding,Deregistration,Network,NetworkInterface,Pipeline,PipelineWorkflow,Process,ResourceReference,Role,RoleBinding,Secret,Subject,System -pkg-path ../../../../api/core/v2 -o ./corev2.gen.graphql
//go:generate go run github.com/sensu/sensu-go/scripts/gen_gqltype -types EntityConfig,EntityState -pkg-path ../../../../api/core/v3 -o ./corev3.gen.graphql
//go:generate go run github.com/sensu/sensu-go/scripts/gengraphql .
