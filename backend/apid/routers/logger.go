package routers

import (
	"github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

var logger = logging.GetLogger("apid.routers").WithFields(logrus.Fields{
	"component": "apid.routers",
})
