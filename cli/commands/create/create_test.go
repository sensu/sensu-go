package create

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"text/template"

	"github.com/ghodss/yaml"
	mockclient "github.com/sensu/sensu-go/cli/client/testing"
	cmdtesting "github.com/sensu/sensu-go/cli/commands/testing"
	corev2 "github.com/sensu/core/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var resourceSpecTmpl = template.Must(template.New("test").Parse(`
{"type": "Check", "spec": {{ .Check }} }
{"type": "Asset", "spec": {{ .Asset }} }
{"type": "Hook", "spec": {{ .Hook }} }
`))

var yamlSpecTmpl = template.Must(template.New("yamltest").Parse(`
type: Check
spec:
  {{ .Check }}
---
type: Asset
spec:
  {{ .Asset }}
---
type: Hook
spec:
  {{ .Hook }}
`))

func mustMarshal(t interface{}) string {
	b, err := json.Marshal(t)
	if err != nil {
		panic(err)
	}
	return string(b)
}

func mustYAMLMarshal(t interface{}) string {
	b, err := yaml.Marshal(t)
	if err != nil {
		panic(err)
	}
	b = bytes.Replace(b, []byte("\n"), []byte("\n  "), -1)
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

var yamlResources = struct {
	Check string
	Asset string
	Hook  string
}{
	Check: mustYAMLMarshal(fixtureCheck),
	Asset: mustYAMLMarshal(fixtureAsset),
	Hook:  mustYAMLMarshal(fixtureHook),
}

func TestCreateCommand(t *testing.T) {
	cli := cmdtesting.NewMockCLI()
	client := cli.Client.(*mockclient.MockClient)
	client.On("PutResource", mock.Anything).Return(nil)
	client.On("PutResource", mock.Anything).Return(nil)
	client.On("PutResource", mock.Anything).Return(nil)

	cmd := CreateCommand(cli)
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

	client.AssertCalled(t, "PutResource", mock.Anything)
	client.AssertCalled(t, "PutResource", mock.Anything)
	client.AssertCalled(t, "PutResource", mock.Anything)
}

func TestCreateCommandYAML(t *testing.T) {
	cli := cmdtesting.NewMockCLI()
	client := cli.Client.(*mockclient.MockClient)
	client.On("PutResource", mock.Anything).Return(nil)
	client.On("PutResource", mock.Anything).Return(nil)
	client.On("PutResource", mock.Anything).Return(nil)

	cmd := CreateCommand(cli)
	td, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(td)

	fp := filepath.Join(td, "input")

	f, err := os.Create(fp)
	require.NoError(t, err)

	err = yamlSpecTmpl.Execute(f, yamlResources)
	require.NoError(t, err)
	require.NoError(t, f.Close())

	require.NoError(t, cmd.Flags().Set("file", fp))
	_, err = cmdtesting.RunCmd(cmd, nil)
	require.NoError(t, err)

	client.AssertCalled(t, "PutResource", mock.Anything)
	client.AssertCalled(t, "PutResource", mock.Anything)
	client.AssertCalled(t, "PutResource", mock.Anything)
}

func TestCreateCommandStdin(t *testing.T) {
	cli := cmdtesting.NewMockCLI()
	client := cli.Client.(*mockclient.MockClient)
	client.On("PutResource", mock.Anything).Return(nil)
	client.On("PutResource", mock.Anything).Return(nil)
	client.On("PutResource", mock.Anything).Return(nil)

	cmd := CreateCommand(cli)
	td, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(td)

	fp := filepath.Join(td, "input")

	f, err := os.Create(fp)
	if err != nil {
		t.Fatal(err)
	}

	if err := yamlSpecTmpl.Execute(f, yamlResources); err != nil {
		t.Fatal(err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		t.Fatal(err)
	}

	defer func(orig *os.File) {
		os.Stdin = orig
	}(os.Stdin)
	os.Stdin = f

	_, err = cmdtesting.RunCmd(cmd, nil)
	require.NoError(t, err)

	client.AssertCalled(t, "PutResource", mock.Anything)
	client.AssertCalled(t, "PutResource", mock.Anything)
	client.AssertCalled(t, "PutResource", mock.Anything)
}
