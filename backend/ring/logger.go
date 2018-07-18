package ring

import "github.com/sirupsen/logrus"

var logger = logrus.WithFields(logrus.Fields{
	"subsystem": "ring",
})
