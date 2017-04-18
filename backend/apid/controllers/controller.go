package controllers

import (
	"github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
)

// A Controller is a grouping of one or more similar HTTP handlers which are
// executed when a route, registered by the Controller, are matched.
type Controller interface {
	Register(*mux.Router)
}

var logger *logrus.Entry

func init() {
	logger = logrus.WithFields(logrus.Fields{
		"component": "apid",
	})
}
