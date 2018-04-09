package migration

import "github.com/sirupsen/logrus"

var logger = logrus.WithFields(logrus.Fields{
	"component": "migration",
})

// Run lauches the migration process
func Run(storeURL string) {
	environments(storeURL)
}
