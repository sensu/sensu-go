package actions

import (
	"github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

var logger = logging.GetLogger("apid").WithFields(logrus.Fields{
	"component": "apid",
})
