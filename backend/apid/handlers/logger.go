package handlers

import (
	"github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

var logger = logging.GetLogger("apid.handlers").WithFields(logrus.Fields{
	"component": "apid.handlers",
})
