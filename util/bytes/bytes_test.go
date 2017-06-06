package bytes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRandom(t *testing.T) {
	b, err := Random(32)
	assert.NoError(t, err)
	assert.Len(t, b, 32)
}
