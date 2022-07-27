package helpers

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/AlecAivazis/survey/v2"
	"github.com/stretchr/testify/assert"
)

func TestDeletePrompt(t *testing.T) {
	assert := assert.New(t)

	// Create temporary files for stdin, stdout & stderr to make it easier to
	// interact with io.
	stdin, err := ioutil.TempFile(os.TempDir(), "sensu-cli-")
	if err != nil {
		t.Fatal("Error creating stdin file: ", stdin.Name())
	}
	defer func() {
		_ = os.Remove(stdin.Name())
	}()

	stdout, err := ioutil.TempFile(os.TempDir(), "sensu-cli-")
	if err != nil {
		t.Fatal("Error creating stdout file: ", stdout.Name())
	}
	defer func() {
		_ = os.Remove(stdout.Name())
	}()

	stderr, err := ioutil.TempFile(os.TempDir(), "sensu-cli-")
	if err != nil {
		t.Fatal("Error creating stderr file: ", stderr.Name())
	}
	defer func() {
		_ = os.Remove(stderr.Name())
	}()

	// TODO: How do we test interactive input
	opts := []survey.AskOpt{
		survey.WithStdio(stdin, stdout, stderr),
	}
	confirmed := ConfirmDeleteWithOpts("test", opts...)
	assert.False(confirmed)
}
