package seeds

import (
	"github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

var logger = logging.GetLogger("seeds").WithFields(logrus.Fields{
	"component": "seeds",
})
