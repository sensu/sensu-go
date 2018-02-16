package importer

import (
	"errors"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/sensu/sensu-go/cli/elements/report"
)

type resourceImporter interface {
	Name() string
	Import(map[string]interface{}) error
	Validate() error
	Save() (int, error)
	SetReporter(*report.Writer)
}

// Importer takes collection of resource importers passes data up when run
type Importer struct {
	AllowWarns bool
	Debug      bool

	importers []resourceImporter
	report    report.Report
}

// NewImporter ...
func NewImporter(s ...resourceImporter) *Importer {
	report := report.New()
	report.Out = os.Stdout
	report.LogLevel = logrus.InfoLevel

	importer := Importer{
		importers: s,
		report:    report,
	}
	return &importer
}

// Run ...
func (i *Importer) Run(d map[string]interface{}) error {
	defer func() { _ = i.report.Flush() }()

	if i.Debug {
		i.report.LogLevel = logrus.DebugLevel
	}

	reporter := report.NewWriter(&i.report)
	for _, r := range i.importers {
		r.SetReporter(&reporter)
	}

	reporter.Debug("instantiating resources for given entries")
	for _, r := range i.importers {
		if err := r.Import(d); err != nil {
			return err
		}
	}

	reporter.Debug("validating given resources")
	for _, r := range i.importers {
		if err := r.Validate(); err != nil {
			return err
		}
	}

	// Guard against saving resources with errors
	if i.report.HasErrors() {
		// TODO: Print how many errors / warnings?
		return errors.New("unable to continue due to errors")
	}

	if !i.AllowWarns && i.report.HasWarnings() {
		return errors.New("please correct any warnings before continuing or use '--force' flag")
	}

	// Save all pending resources
	for _, r := range i.importers {
		if n, err := r.Save(); n > 0 && err == nil {
			reporter.Infof("all '%s' resources imported", r.Name())
		}
	}

	return nil
}
