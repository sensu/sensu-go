package create

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/ghodss/yaml"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	mockclient "github.com/sensu/sensu-go/cli/client/testing"
	cmdtesting "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/sensu/sensu-go/types"
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
	fixtureCheck = types.FixtureCheck("foo")
	fixtureAsset = types.FixtureAsset("bar")
	fixtureHook  = types.FixtureHook("baz")
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

func TestValidateResources(t *testing.T) {
	tests := []struct {
		name          string
		resource      *types.Wrapper
		namespace     string
		wantNamespace string
	}{
		{
			name: "a namespaced resource with a configured namespace should not be modified",
			resource: &types.Wrapper{
				ObjectMeta: corev2.NewObjectMeta("check-cpu", "default"),
				Value: &corev2.CheckConfig{
					ObjectMeta: corev2.NewObjectMeta("check-cpu", "default"),
				},
			},
			namespace:     "dev",
			wantNamespace: "default",
		},
		{
			name: "a namespaced resource without a configured namespace should use the provided namespace",
			resource: &types.Wrapper{
				ObjectMeta: corev2.NewObjectMeta("check-cpu", ""),
				Value: &corev2.CheckConfig{
					ObjectMeta: corev2.NewObjectMeta("check-cpu", ""),
				},
			},
			namespace:     "dev",
			wantNamespace: "dev",
		},
		{
			name: "a global resource should not have a namespace configured",
			resource: &types.Wrapper{
				ObjectMeta: corev2.NewObjectMeta("admin-role", ""),
				Value: &corev2.ClusterRole{
					ObjectMeta: corev2.NewObjectMeta("admin-role", ""),
				},
			},
			namespace:     "dev",
			wantNamespace: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resources := []*types.Wrapper{tt.resource}
			_ = ValidateResources(resources, tt.namespace)

			if tt.resource.ObjectMeta.Namespace != tt.wantNamespace {
				t.Errorf("ValidateResources() wrapper namespace = %q, want namespace %q", tt.resource.ObjectMeta.Namespace, tt.wantNamespace)
			}
			if tt.resource.Value != nil && tt.resource.Value.GetObjectMeta().Namespace != tt.wantNamespace {
				t.Errorf("ValidateResources() wrapper's resource namespace = %q, want namespace %q", tt.resource.Value.GetObjectMeta().Namespace, tt.wantNamespace)
			}
		})
	}
}

func TestValidateResourcesStderr(t *testing.T) {
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	ch := make(chan string)
	// copy the output in a separate goroutine so printing can't block indefinitely
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		ch <- buf.String()
	}()

	resources := []*types.Wrapper{&types.Wrapper{
		ObjectMeta: corev2.NewObjectMeta("check-cpu", "default"),
	}}
	_ = ValidateResources(resources, "default")

	// Reset stderr
	w.Close()
	os.Stderr = oldStderr

	errMsg := <-ch
	errMsg = strings.TrimSpace(errMsg)
	wantErr := `error validating resource #0 with name "check-cpu" and namespace "default": resource is nil`
	if errMsg != wantErr {
		t.Errorf("ValidateResources() err = %s, want %s", errMsg, wantErr)
	}
}
