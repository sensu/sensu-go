package middlewares

import "github.com/sirupsen/logrus"

var Logger = logrus.New().WithFields(logrus.Fields{
	"component": "apid",
})

func init() {
	Logger.Logger.SetFormatter(&logrus.JSONFormatter{})
}
