package api

import (
	"github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

var logger = logging.GetLogger("backend.api").WithFields(logrus.Fields{
	"component": "backend.api",
})
