package schedulerd

import "github.com/Sirupsen/logrus"

var logger = logrus.WithFields(logrus.Fields{
	"component": "schedulerd",
})
