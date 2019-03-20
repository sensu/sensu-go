package v2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateFrequency(t *testing.T) {
	assert.Error(t, ValidateFrequency(UpperBound+1))
	assert.Error(t, ValidateFrequency(LowerBound-1))
	assert.NoError(t, ValidateFrequency(UpperBound-1))
	assert.NoError(t, ValidateFrequency(LowerBound+1))
}
