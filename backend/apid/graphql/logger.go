package graphql

import (
	"github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

var logger = logging.GetLogger("apid.graphql").WithFields(logrus.Fields{
	"component": "apid.graphql",
})
