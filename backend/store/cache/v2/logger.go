package v2

import "github.com/sirupsen/logrus"

var logger = logrus.WithFields(logrus.Fields{
	"component":   "cache",
	"api_version": "v2",
})
