package handler

import (
	"github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

var logger = logging.GetLogger("pipeline.legacy").WithFields(logrus.Fields{
	"component": "pipeline.legacy",
})
