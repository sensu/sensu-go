package keepalived

import (
	"github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

var logger = logging.GetLogger("keepalived").WithFields(logrus.Fields{
	"component": "keepalived",
})
