package rbac

import (
	"github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

var logger = logging.GetLogger("rbac").WithFields(logrus.Fields{
	"component": "rbac",
})
