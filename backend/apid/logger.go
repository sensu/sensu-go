package apid

import (
	"github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

func getLogger() *logrus.Logger {
	return logging.GetLogger("backend.apid")
}
