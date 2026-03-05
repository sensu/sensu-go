package tessend

import (
	"github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

var logger = logging.GetLogger("tessend").WithFields(logrus.Fields{
	"component": "tessend",
})
