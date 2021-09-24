package metrics

import "github.com/sirupsen/logrus"

func LogError(logger *logrus.Entry, name string, err error) {
	fields := logrus.Fields{
		"name":  name,
		"error": err,
	}
	logger.WithFields(fields).Errorf("failed to register metric")
}
