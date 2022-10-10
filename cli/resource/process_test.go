package resource

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/types/compat"
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

func TestManagedByLabelPutter_label(t *testing.T) {
	tests := []struct {
		name     string
		resource types.Wrapper
		want     map[string]string
	}{
		{
			name: "the label is set on both the inner and outer labels",
			resource: types.WrapResource(&corev2.CheckConfig{
				ObjectMeta: corev2.ObjectMeta{
					Name: "foo",
				},
			}),
			want: map[string]string{
				corev2.ManagedByLabel: "sensuctl",
			},
		},
		{
			name: "the label is appended to an existing list of labels",
			resource: types.WrapResource(&corev2.CheckConfig{
				ObjectMeta: corev2.ObjectMeta{
					Labels: map[string]string{
						"region": "us-west-2",
					},
				},
			}),
			want: map[string]string{
				"region":              "us-west-2",
				corev2.ManagedByLabel: "sensuctl",
			},
		},
		{
			name: "the label does not overwrites the sensu-agent value",
			resource: types.WrapResource(&corev2.CheckConfig{
				ObjectMeta: corev2.ObjectMeta{
					Labels: map[string]string{
						corev2.ManagedByLabel: "sensu-agent",
					},
				},
			}),
			want: map[string]string{
				corev2.ManagedByLabel: "sensu-agent",
			},
		},
		{
			name: "the label overwrites any other existing value",
			resource: types.WrapResource(&corev2.CheckConfig{
				ObjectMeta: corev2.ObjectMeta{
					Labels: map[string]string{
						corev2.ManagedByLabel: "web-ui",
					},
				},
			}),
			want: map[string]string{
				corev2.ManagedByLabel: "sensuctl",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.resource
			processor := NewManagedByLabelPutter("sensuctl")
			processor.label(&got)

			if !reflect.DeepEqual(got.ObjectMeta.Labels, tt.want) {
				t.Errorf("inner labels = %v, want %v", got.ObjectMeta.Labels, tt.want)
			}
			if !reflect.DeepEqual(compat.GetObjectMeta(got.Value).Labels, tt.want) {
				t.Errorf("outer labels = %v, want %v", compat.GetObjectMeta(got.Value).Labels, tt.want)
			}
		})
	}
}
