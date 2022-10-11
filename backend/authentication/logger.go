package authentication

import "github.com/sirupsen/logrus"

var Logger = logrus.New().WithFields(logrus.Fields{
	"component": "authentication",
})

func init() {
	Logger.Logger.SetFormatter(&logrus.JSONFormatter{})
}
