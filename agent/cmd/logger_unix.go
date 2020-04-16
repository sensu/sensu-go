// +build !windows

package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

func init() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGUSR1)
	go func() {
		for range sigs {
			level := logrus.GetLevel()
			newLevel := logging.IncrementLogLevel(level)
			logrus.Warnf("set log level to %s", newLevel)
			logrus.SetLevel(newLevel)
			if newLevel == logrus.WarnLevel {
				// repeat the log call, as it wouldn't have been logged at
				// error level.
				logrus.Warnf("set log level to %s", newLevel)
			}
		}
	}()
}
