package helpers

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeletePrompt(t *testing.T) {
	assert := assert.New(t)

	// TODO: How do we test interactive input
	confirmed := ConfirmDelete("test")

	// Print a newline after running the command as a workaround for a bug in
	// test2json. See https://github.com/golang/go/issues/38063 for more
	// information.
	fmt.Fprintf(os.Stdout, "\n")

	assert.False(confirmed)
}
