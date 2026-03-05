package filter

import (
	"github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

var logger = logging.GetLogger("pipeline.filter").WithFields(logrus.Fields{
	"component": "pipeline.filter",
})
