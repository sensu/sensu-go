package schedulerd

import (
	"github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

var logger = logging.GetLogger("schedulerd").WithFields(logrus.Fields{
	"component": "schedulerd",
})
