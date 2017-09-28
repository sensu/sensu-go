package importer

import (
	"fmt"
	"testing"

	"github.com/sensu/sensu-go/cli/elements/report"
	"github.com/stretchr/testify/assert"
)

type TestImporter struct {
	MyName       string
	EmitWarnings bool
	EmitErrors   bool

	reporter report.Writer
}

func (t *TestImporter) Name() string {
	return t.MyName
}

func (t *TestImporter) SetReporter(r *report.Writer) {
	t.reporter = *r
}

func (t *TestImporter) Import(map[string]interface{}) error {
	return nil
}

func (t *TestImporter) Validate() error {
	if t.EmitErrors {
		t.reporter.Error("OMG")
	}

	if t.EmitWarnings {
		t.reporter.Warn("slightly less OMG!")
	}

	return nil
}

func (t *TestImporter) Save() (int, error) {
	return 60, nil
}

func TestImportRunner(t *testing.T) {
	importer := TestImporter{MyName: "no errors or warns"}
	warnImporter := TestImporter{MyName: "has warns no errors", EmitWarnings: true}
	errImporter := TestImporter{MyName: "has errors no warns", EmitErrors: true}
	badImporter := TestImporter{MyName: "has many issues", EmitErrors: true, EmitWarnings: true}

	testCases := []struct {
		importer   resourceImporter
		allowWarns bool
		wantError  bool
	}{
		{&importer, true, false},
		{&importer, false, false},
		{&warnImporter, true, false},
		{&warnImporter, false, true},
		{&errImporter, true, true},
		{&errImporter, false, true},
		{&badImporter, true, true},
		{&badImporter, false, true},
	}
	for i, tc := range testCases {
		t.Run(
			fmt.Sprintf("Run importer that %s", tc.importer.Name(), tc.allowWarns),
			func(t *testing.T) {
				assert := assert.New(t)
				runner := NewImporter(tc.importer)
				runner.AllowWarns = tc.allowWarns
				runner.Debug = (i % 2) == 0

				if got := runner.Run(map[string]interface{}{}); tc.wantError {
					assert.Error(got)
				} else {
					assert.NoError(got)
				}
			},
		)
	}
}
