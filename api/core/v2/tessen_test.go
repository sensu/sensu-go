package v2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateInterval(t *testing.T) {
	assert.Error(t, ValidateInterval(UpperBound+1))
	assert.Error(t, ValidateInterval(LowerBound-1))
	assert.NoError(t, ValidateInterval(UpperBound-1))
	assert.NoError(t, ValidateInterval(LowerBound+1))
}
