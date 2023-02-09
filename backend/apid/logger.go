package apid

import "github.com/sirupsen/logrus"

var Logger = logrus.New().WithFields(logrus.Fields{
	"component": "apid",
})

func init() {
	Logger.Logger.SetFormatter(&logrus.JSONFormatter{})

	// There are other functions and fields that could be of interest. We could
	// allow their configuration too, for example to have different component
	// spit out their log in different files:
	// - SetNoLock()
	// - Out
	// - SetOutput(output io.Writer)
}
