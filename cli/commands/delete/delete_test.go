package delete

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"text/template"

	corev2 "github.com/sensu/core/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	mockclient "github.com/sensu/sensu-go/cli/client/testing"
	cmdtesting "github.com/sensu/sensu-go/cli/commands/testing"
)

var resourceSpecTmpl = template.Must(template.New("test").Parse(`
{"type": "Check", "spec": {{ .Check }} }
{"type": "Asset", "spec": {{ .Asset }} }
{"type": "Hook", "spec": {{ .Hook }} }
`))

func mustMarshal(t interface{}) string {
	b, err := json.Marshal(t)
	if err != nil {
		panic(err)
	}
	return string(b)
}

var (
	fixtureCheck = corev2.FixtureCheck("foo")
	fixtureAsset = corev2.FixtureAsset("bar")
	fixtureHook  = corev2.FixtureHook("baz")
)

var resources = struct {
	Check string
	Asset string
	Hook  string
}{
	Check: mustMarshal(fixtureCheck),
	Asset: mustMarshal(fixtureAsset),
	Hook:  mustMarshal(fixtureHook),
}

func TestDeleteCommand(t *testing.T) {
	cli := cmdtesting.NewMockCLI()
	client := cli.Client.(*mockclient.MockClient)
	client.On("Delete", mock.Anything).Return(nil)
	client.On("Delete", mock.Anything).Return(nil)
	client.On("Delete", mock.Anything).Return(nil)

	cmd := DeleteCommand(cli)
	td, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(td)

	fp := filepath.Join(td, "input")

	f, err := os.Create(fp)
	require.NoError(t, err)

	err = resourceSpecTmpl.Execute(f, resources)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	require.NoError(t, cmd.Flags().Set("file", fp))
	_, err = cmdtesting.RunCmd(cmd, nil)
	require.NoError(t, err)

	client.AssertCalled(t, "Delete", mock.Anything)
	client.AssertCalled(t, "Delete", mock.Anything)
	client.AssertCalled(t, "Delete", mock.Anything)
}
