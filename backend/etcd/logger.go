package etcd

import (
	"github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

var logger = logging.GetLogger("etcd").WithFields(logrus.Fields{
	"component": "etcd",
})
