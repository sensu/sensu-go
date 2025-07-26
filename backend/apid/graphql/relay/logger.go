package relay

import (
	"github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

func getLogger() *logrus.Logger {
	return logging.GetLogger("graphql.relay")
}
