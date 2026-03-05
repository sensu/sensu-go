package util_api

import (
	"github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

func getLogger() *logrus.Logger {
	return logging.GetLogger("apid.graphql.util.api")
}
