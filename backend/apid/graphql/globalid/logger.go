package globalid

import (
	"github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

var defaultLogger = logging.GetLogger("apid.graphql.globalid").WithFields(logrus.Fields{
	"component": "apid.graphql.globalid",
})
