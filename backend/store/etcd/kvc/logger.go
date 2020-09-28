package kvc

import "github.com/sirupsen/logrus"

var logger = logrus.WithFields(logrus.Fields{
	"component": "sensu.store.etcd.kvc",
})
