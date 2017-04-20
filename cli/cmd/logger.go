package main

import "github.com/Sirupsen/logrus"

var logger = logrus.WithFields(logrus.Fields{
	"component": "cmd",
})

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{})
}
