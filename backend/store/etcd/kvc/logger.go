package kvc

import (
	"github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

var logger = logging.GetLogger("sensu.store.etcd.kvc").WithFields(logrus.Fields{
	"component": "sensu.store.etcd.kvc",
})
