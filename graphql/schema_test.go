package graphqlschema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixtureUser(t *testing.T) {
	assert := assert.New(t)
	schema := Schema()

	assert.NotNil(schema.QueryType(), "query type has been configured")
	// assert.NotNil(schema.MutationType(), "mutation type has been configured")
	assert.NotEmpty(schema.Directives(), "directives are present on schema")
	assert.NotEmpty(schema.TypeMap(), "types have been added to the schema")
}
