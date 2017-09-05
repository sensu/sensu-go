package importer

import (
	"errors"
	"os"

	"github.com/Sirupsen/logrus"
)

type resourceImporter interface {
	Name() string
	Import(map[string]interface{}) error
	Validate() error
	Save() (int, error)
	SetReporter(*Reporter)
}

// Importer takes collection of resource importers passes data up when run
type Importer struct {
	AllowWarns bool
	Debug      bool

	importers []resourceImporter
	reporter  Reporter
}

// NewImporter ...
func NewImporter(s ...resourceImporter) *Importer {
	importer := Importer{importers: s}

	importer.reporter.LogLevel = logrus.InfoLevel
	importer.reporter.Out = os.Stderr

	return &importer
}

// Run ...
func (i *Importer) Run(d map[string]interface{}) error {
	defer i.reporter.Flush()

	if i.Debug {
		i.reporter.LogLevel = logrus.DebugLevel
	}

	for _, r := range i.importers {
		r.SetReporter(&i.reporter)
	}

	// Instantiate resources for each given entry
	for _, r := range i.importers {
		r.Import(d)
	}

	// Validate all resources
	for _, r := range i.importers {
		r.Validate()
	}

	// Guard against saving resources with errors
	if i.reporter.hasErrors() {
		// TODO: Print how many errors / warnings?
		return errors.New("unable to continue due errors")
	}

	if !i.AllowWarns && i.reporter.hasWarnings() {
		return errors.New("please correct any warnings before continuing or use '--force' flag")
	}

	// Save all pending resources
	for _, r := range i.importers {
		if n, err := r.Save(); n > 0 && err == nil {
			i.reporter.Infof("all '%s' imported", r.Name())
		}
	}

	return nil
}
