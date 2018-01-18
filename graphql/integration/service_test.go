package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/graphql-go/graphql/language/ast"
	"github.com/sensu/sensu-go/graphql"
	schema "github.com/sensu/sensu-go/graphql/integration/testdata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExerciseService(t *testing.T) {
	svc := graphql.NewService()
	require.NotNil(t, svc)

	schema.RegisterFoo(svc, &fooImpl{})
	schema.RegisterQueryRoot(svc, &queryRootImpl{})
	schema.RegisterBar(svc, &exResolver{})
	schema.RegisterUrl(svc, &urlHandler{})
	schema.RegisterInputType(svc)
	schema.RegisterFeed(svc, &exResolver{})
	schema.RegisterSite(svc)
	schema.RegisterLocale(svc)
	schema.RegisterSchema(svc)

	err := svc.Regenerate()
	require.NoError(t, err)

	ctx := context.Background()
	res := svc.Do(ctx, "query { myBar { one } }", map[string]interface{}{})
	require.Empty(t, res.Errors)
	assert.NotEmpty(t, res.Data)
}

type exResolver struct{}

func (*exResolver) IsTypeOf(_ interface{}, _ graphql.IsTypeOfParams) bool {
	return true
}
func (*exResolver) ResolveType(_ interface{}, _ graphql.ResolveTypeParams) *graphql.Type {
	return &schema.FooType
}

type fooImpl struct {
	schema.FooAliases
	exResolver
}

type queryRootImpl struct {
	schema.QueryRootAliases
	exResolver
}

func (*queryRootImpl) MyBar(_ graphql.ResolveParams) (interface{}, error) {
	return map[string]interface{}{"one": 1}, nil
}

type urlHandler struct{}

func (*urlHandler) Serialize(x interface{}) interface{} {
	return fmt.Sprintf("https://sensu.io/%v", x)
}
func (*urlHandler) ParseValue(interface{}) interface{} {
	return nil
}
func (*urlHandler) ParseLiteral(ast.Value) interface{} {
	return nil
}
