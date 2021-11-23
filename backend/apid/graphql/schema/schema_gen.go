package schema

//go:generate go run github.com/sensu/sensu-go/scripts/gen_gqltype -types Secret -pkg-path ../../../../api/core/v2 -o ./corev2.gen.graphql
//go:generate go run github.com/sensu/sensu-go/scripts/gengraphql .
