package resource

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/cli"
	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/client/config"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockProcessor struct {
	mock.Mock
}

func (m *mockProcessor) Process(client client.GenericClient, resources []*types.Wrapper) error {
	return m.Called(client, resources).Error(0)
}

func TestProcessURL(t *testing.T) {
	bytes, err := json.Marshal(types.WrapResource(corev2.FixtureCheck("check")))
	assert.NoError(t, err)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(bytes)
		assert.NoError(t, err)
	}))
	defer ts.Close()

	processor := &mockProcessor{}
	processor.On("Process", mock.Anything, mock.Anything).Return(nil)
	config := &config.MockConfig{}
	config.On("Namespace").Return("")

	urly, err := url.Parse(ts.URL)
	assert.NoError(t, err)
	assert.NoError(t, ProcessURL(&cli.SensuCli{Config: config}, ts.Client(), urly, ts.URL, false, processor))
}

func TestProcessFile(t *testing.T) {
	td, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(td)

	fp := filepath.Join(td, "input")
	err = ioutil.WriteFile(fp, []byte(`{"type": "Namespace", "spec": {"name": "foo"}}`), 0644)
	assert.NoError(t, err)

	processor := &mockProcessor{}
	processor.On("Process", mock.Anything, mock.Anything).Return(nil)
	config := &config.MockConfig{}
	config.On("Namespace").Return("")
	assert.NoError(t, ProcessFile(&cli.SensuCli{Config: config}, fp, false, processor))
}
