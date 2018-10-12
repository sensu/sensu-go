package agent

import (
	"github.com/echlebek/crock"
	time "github.com/echlebek/timeproxy"
)

var mockTime = crock.NewTime(time.Unix(0, 0))

func init() {
	// Resolution and Multiplier make the mock time advance every `Resolution`
	// by `Resolution * Multiplier`. That is, for the values defined here:
	// advance the mock time by 100ms every 1ms of real time, giving us a
	// precision of about 100ms for the mock time.
	mockTime.Resolution = time.Millisecond
	mockTime.Multiplier = 100
	time.TimeProxy = mockTime
}
