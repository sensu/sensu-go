package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeletePrompt(t *testing.T) {
	assert := assert.New(t)
	out := exWriter{}

	// TODO: How do we test interactive input
	confirmed := ConfirmDelete("test", &out)
	assert.False(confirmed)
	assert.Contains(out.result, "Are you sure")
}

type exWriter struct {
	result string
}

func (w *exWriter) Clean() {
	w.result = ""
}

func (w *exWriter) Write(p []byte) (int, error) {
	w.result += string(p)
	return 0, nil
}
