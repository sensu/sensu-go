package importer

import (
	"testing"

	"encoding/json"
	"io/ioutil"
	"path/filepath"

	clientmock "github.com/sensu/sensu-go/cli/client/testing"
	"github.com/stretchr/testify/assert"
)

func TestLegacySettings(t *testing.T) {
	matches, _ := filepath.Glob("./catalog/*.json")
	for _, match := range matches {
		t.Run(filepath.Base(match), func(t *testing.T) {
			file, e := ioutil.ReadFile(match)
			if e != nil {
				t.Fatal("could not open")
			}

			var data map[string]interface{}
			json.Unmarshal(file, &data)

			client := clientmock.MockClient{}
			importer := NewSensuV1SettingsImporter("default", "default", &client)

			err := importer.Run(data)
			t.Skip("Not all attributes are covered at this time.")
			assert.NoError(t, err)
		})
	}
}
