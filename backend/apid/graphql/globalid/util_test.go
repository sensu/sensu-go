package globalid

import (
	"testing"

	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestStandardDecoder(t *testing.T) {
	assert := assert.New(t)

	handler := types.FixtureHandler("myHandler")
	encoderFn := standardEncoder("handlers", "Name")
	components := encoderFn(handler)

	assert.Equal("handlers", components.Resource())
	assert.Equal("default", components.Namespace())
	assert.Equal("myHandler", components.UniqueComponent())
}
