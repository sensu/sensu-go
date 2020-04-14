package cmd

import (
	"github.com/sirupsen/logrus"
)

var logger = logrus.WithFields(logrus.Fields{
	"component": "cmd",
})

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
}
