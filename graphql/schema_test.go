package graphqlschema

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFixtureUser(t *testing.T) {
	assert := assert.New(t)
	schema := Schema()

	if schema.Directive("query") == nil {
		assert.FailNow("query directive not present.")
	}
}
