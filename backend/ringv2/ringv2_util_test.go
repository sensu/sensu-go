package ringv2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathUnPath(t *testing.T) {
	assert := assert.New(t)

	expectedNs := "default"
	expectedSub := "some-subscription"

	path := Path(expectedNs, expectedSub)
	obtainedNs, obtainedSub, err := UnPath(path)
	assert.NoError(err)
	assert.Equal(expectedNs, obtainedNs)
	assert.Equal(expectedSub, obtainedSub)
}
