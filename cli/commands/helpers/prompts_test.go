package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeletePrompt(t *testing.T) {
	assert := assert.New(t)

	// TODO: How do we test interactive input
	confirmed := ConfirmDelete("test")
	assert.False(confirmed)
}
