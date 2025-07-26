package cmd

import (
	"github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

var logger = logging.GetLogger("cmd").WithFields(logrus.Fields{
	"component": "cmd",
})
