package authentication

import (
	"github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

var logger = logging.GetLogger("authentication").WithFields(logrus.Fields{
	"component": "authentication",
})
