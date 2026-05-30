package secrets

import (
	"github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

var logger = logging.GetLogger("secrets").WithFields(logrus.Fields{
	"component": "secrets",
})
