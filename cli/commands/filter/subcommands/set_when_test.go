package subcommands

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	v2 "github.com/sensu/core/v2"
	client "github.com/sensu/sensu-go/cli/client/testing"
	stest "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func fileFromString(t *testing.T, s string) (string, *os.File, func()) {
	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	name := filepath.Join(dir, "timewindows.json")
	tf, err := os.Create(name)
	require.NoError(t, err)
	cleanup := func() {
		_ = tf.Close()
		assert.NoError(t, os.RemoveAll(dir))
	}
	_, err = fmt.Fprintln(tf, s)
	require.NoError(t, err)
	require.NoError(t, tf.Sync())
	_, err = tf.Seek(0, 0)
	require.NoError(t, err)
	return name, tf, cleanup
}

func TestSetWhenCommand(t *testing.T) {
	const timeWindowsJSON = `{"days":{"all":[{"begin":"3:00 PM","end":"4:00 PM"}]}}`
	tests := []struct {
		args           []string
		useflag        bool
		stdin          string
		fetchResponse  error
		updateResponse error
		expectedOutput string
		expectError    bool
	}{
		{[]string{}, false, "", nil, nil, "Usage", true},
		{[]string{"foo"}, false, "", errors.New("error"), nil, "", true},
		{[]string{"bar"}, false, "", nil, errors.New("error"), "", true},
		{[]string{"filter1"}, false, "", nil, nil, "", true},
		{[]string{"filter1"}, false, timeWindowsJSON, nil, nil, "Updated", false},
		{[]string{"filter1"}, false, "invalidjson", nil, nil, "", true},
		{[]string{"filter1"}, true, timeWindowsJSON, nil, nil, "", false},
	}

	for i, test := range tests {
		name := ""
		if len(test.args) > 0 {
			name = test.args[0]
		}
		t.Run(fmt.Sprintf("test %d", i), func(t *testing.T) {
			filter := v2.FixtureEventFilter("filter1")
			cli := stest.NewMockCLI()
			client := cli.Client.(*client.MockClient)
			client.On("FetchFilter", name).Return(filter, test.fetchResponse)
			client.On("UpdateFilter", mock.Anything).Return(test.updateResponse)
			cmd := SetWhenCommand(cli)
			name, stdin, cleanup := fileFromString(t, test.stdin)
			defer cleanup()
			if test.useflag {
				require.NoError(t, stdin.Close())
				require.NoError(t, cmd.Flags().Set("file", name))
			} else {
				os.Stdin = stdin
			}
			out, err := stest.RunCmd(cmd, test.args)
			if test.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Regexp(t, test.expectedOutput, out)
		})
	}
}
