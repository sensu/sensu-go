package actions

import "github.com/sirupsen/logrus"

var logger = logrus.WithFields(logrus.Fields{
	"component": "apid",
})
