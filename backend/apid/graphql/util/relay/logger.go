package util_relay

import "github.com/sirupsen/logrus"

var logger = logrus.WithFields(logrus.Fields{
	"component": "apid.graphql.util.relay",
})
