package transport

import (
	"github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

var logger = logging.GetLogger("transport").WithFields(logrus.Fields{
	"component": "transport",
})
