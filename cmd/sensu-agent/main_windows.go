package main

import (
	"github.com/sensu/sensu-go/agent/cmd"
	"github.com/sirupsen/logrus"
)

var logger = logrus.WithFields(logrus.Fields{
	"component": "agent",
})

func main() {
	if err := cmd.Execute(); err != nil {
		logger.WithError(err).Fatal("error executing sensu-agent")
	}
}
