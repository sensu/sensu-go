package assetmanager

import "github.com/Sirupsen/logrus"

var logger *logrus.Entry

func init() {
	logger = logrus.WithFields(logrus.Fields{
		"component": "assetmanager",
	})
}
