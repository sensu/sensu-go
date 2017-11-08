package graphqlschema

import (
	"encoding/json"
	"testing"

	"github.com/graphql-go/graphql/testutil"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/context"
)

func TestExecute(t *testing.T) {
	assert := assert.New(t)

	// Ensure that introspection query can be executed
	result := Execute(context.TODO(), testutil.IntrospectionQuery, nil)
	assert.NotNil(result)
	assert.NotEmpty(result.Data)
	assert.Empty(result.Errors)

	// Ensure that given result can be serialized
	serializedResult, err := json.Marshal(result)
	assert.NoError(err)
	assert.NotEmpty(serializedResult)
}
