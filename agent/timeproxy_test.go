package agent

import (
	"github.com/echlebek/crock"
	time "github.com/echlebek/timeproxy"
)

var mockTime = crock.NewTime(time.Unix(0, 0))

func init() {
	mockTime.Resolution = time.Microsecond
	mockTime.Multiplier = 5000
	time.TimeProxy = mockTime
}
