package leader

import (
	"sync/atomic"
	"time"

	"github.com/Sirupsen/logrus"
)

var (
	logger      = logrus.WithFields(logrus.Fields{"package": "leader"})
	logInterval = int64(time.Minute)
)

// SetLogInterval sets the interval for logging from this package. Values <= 0
// are ignored. SetLogInterval has no effect after Init is run.
func SetLogInterval(t time.Duration) {
	if t > 0 {
		atomic.StoreInt64(&logInterval, int64(t))
	}
}

func getLogInterval() time.Duration {
	return time.Duration(atomic.LoadInt64(&logInterval))
}

func logPeriodic(nodeName, leaderName string, totalWork int64) {
	isLeader := nodeName == leaderName
	logger.WithFields(logrus.Fields{
		"leading":        isLeader,
		"node_name":      nodeName,
		"leader_name":    leaderName,
		"work_completed": totalWork,
	}).Info("leader")
}
