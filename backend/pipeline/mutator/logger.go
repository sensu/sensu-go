package mutator

import "github.com/sirupsen/logrus"

var logger = logrus.WithFields(logrus.Fields{
	"component": "pipeline/legacy",
})
