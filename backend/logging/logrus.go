package logging

import (
	"github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

var logger = logging.GetLogger("backend").WithFields(logrus.Fields{
	"component": "backend",
})
