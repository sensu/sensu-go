package pipelined

import (
	"github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

var logger = logging.GetLogger("pipelined").WithFields(logrus.Fields{
	"component": "pipelined",
})
