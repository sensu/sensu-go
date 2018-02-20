package importer

import (
	"fmt"
	"testing"

	"encoding/json"
	"io/ioutil"
	"path/filepath"

	clientmock "github.com/sensu/sensu-go/cli/client/testing"
	"github.com/sensu/sensu-go/cli/elements/report"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestLegacyFilterImporter(t *testing.T) {
	testCases := []struct {
		name           string
		data           []byte
		expectedFilter types.EventFilter
	}{
		{
			"simple attribute",
			[]byte(`"attributes": {"foo": "bar"}`),
			types.EventFilter{
				Action:     types.EventFilterActionAllow,
				Statements: []string{"event.Foo == 'bar'"},
			},
		},
		{
			"client attribute",
			[]byte(`"attributes": {"client": {"environment": "production"}}`),
			types.EventFilter{
				Name:       "foo",
				Action:     types.EventFilterActionAllow,
				Statements: []string{"event.Entity.Environment == 'production'"},
			},
		},
		{
			"check attribute",
			[]byte(`"attributes": {"check": {"type": "metric"}}`),
			types.EventFilter{
				Name:       "foo",
				Action:     types.EventFilterActionAllow,
				Statements: []string{"event.Check.Type == 'metric'"},
			},
		},
		{
			"negate attribute",
			[]byte(`"attributes": {"client": {"environment": "production"}}, "negate": true`),
			types.EventFilter{
				Name:       "foo",
				Action:     types.EventFilterActionDeny,
				Statements: []string{"event.Entity.Environment == 'production'"},
			},
		},
		{
			"simple int attribute",
			[]byte(`"attributes": {"check": {"interval": 30}}`),
			types.EventFilter{
				Action:     types.EventFilterActionAllow,
				Statements: []string{"event.Check.Interval == 30"},
			},
		},
		{
			"simple boolean attribute",
			[]byte(`"attributes": {"check": {"standalone": true}}`),
			types.EventFilter{
				Action:     types.EventFilterActionAllow,
				Statements: []string{"event.Check.Standalone == true"},
			},
		},
		{
			"invalid eval attribute",
			[]byte(`"attributes": {"occurrences": "eval: value == 1", "foo": "bar"}`),
			types.EventFilter{
				Action:     types.EventFilterActionAllow,
				Statements: []string{"event.Foo == 'bar'"},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(
			fmt.Sprintf("Run filter importer with %s", tc.name), func(t *testing.T) {
				// Make sure the raw JSON correspond to the 1.x spec
				rawData := []byte(`{"filters": {"test": {`)
				rawData = append(rawData, tc.data...)
				rawData = append(rawData, []byte(`}}}`)...)

				// Import the raw JSON data
				var data map[string]interface{}
				if err := json.Unmarshal(rawData, &data); err != nil {
					assert.FailNow(t, err.Error())
				}

				// Create the filter importer
				client := clientmock.MockClient{}
				filterImporter := &LegacyFilterImporter{
					Org:      "default",
					Env:      "default",
					SaveFunc: client.CreateFilter,
				}
				i := NewImporter(filterImporter)

				// Set the reporter
				defer func() { _ = i.report.Flush() }()
				reporter := report.NewWriter(&i.report)
				i.importers[0].SetReporter(&reporter)

				if err := i.importers[0].Import(data); err != nil {
					assert.FailNow(t, err.Error())
				}

				assert.Equal(t, tc.expectedFilter.Action, filterImporter.filters[0].Action)
				assert.Equal(t, tc.expectedFilter.Statements, filterImporter.filters[0].Statements)
			},
		)
	}
}

func TestLegacySettings(t *testing.T) {
	matches, _ := filepath.Glob("./catalog/*.json")
	for _, match := range matches {
		t.Run(filepath.Base(match), func(t *testing.T) {
			file, e := ioutil.ReadFile(match)
			if e != nil {
				t.Fatal("could not open")
			}

			var data map[string]interface{}
			_ = json.Unmarshal(file, &data)

			client := clientmock.MockClient{}
			importer := NewSensuV1SettingsImporter("default", "default", &client)

			err := importer.Run(data)
			t.Skip("Not all attributes are covered at this time.")
			assert.NoError(t, err)
		})
	}
}
