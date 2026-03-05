package cache

import (
	"github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

var logger = logging.GetLogger("cache").WithFields(logrus.Fields{
	"component":     "cache",
	"cache_version": "v1",
})
