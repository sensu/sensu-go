//go:build !windows
// +build !windows

package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sensu/sensu-go/util/logging"
	"github.com/sirupsen/logrus"
)

// This needs to be adapted, or scrapped altogether in favor of some form of
// live configuration reloading. If we try to keep this mechanism, we are in a
// catch 22: if we increment all the loggers by one from their current level
// we'll end up in a mess of levels, and if we set all the loggers to the same
// global value we loose the differenciated log levels until the next restart.
//
// Many Unix daemons like sshd and unbound implement graceful configuration
// reloading to circumvent this issue: SIGHUP is used to tell the process to
// reread the configuration file it was started with and adjust the
// configuration accordingly (for some compatible configuration parameters, not
// all), typically without dropping active connections.
// See: https://en.wikipedia.org/wiki/Signal_(IPC)#SIGHUP
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
