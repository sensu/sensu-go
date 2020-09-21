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
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestProcessURL(t *testing.T) {
	bytes, err := json.Marshal(types.WrapResource(corev2.FixtureCheck("check")))
	assert.NoError(t, err)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write(bytes)
		assert.NoError(t, err)
	}))
	defer ts.Close()

	urly, err := url.Parse(ts.URL)
	assert.NoError(t, err)
	_, err = ProcessURL(ts.Client(), urly, ts.URL, false)
	assert.NoError(t, err)
}

func TestProcessFile(t *testing.T) {
	td, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(td)

	fp := filepath.Join(td, "input")
	err = ioutil.WriteFile(fp, []byte(`{"type": "Namespace", "spec": {"name": "foo"}}`), 0644)
	assert.NoError(t, err)

	_, err = ProcessFile(fp, false)
	assert.NoError(t, err)
}
