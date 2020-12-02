package transformers

import (
	"github.com/echlebek/crock"
	time "github.com/echlebek/timeproxy"
)

var mockTime = crock.NewTime(time.Unix(1257894000, 0))

func init() {
	// Resolution and Multiplier make the mock time advance every `Resolution`
	// by `Resolution * Multiplier`. That is, for the values defined here:
	// advance the mock time by 500ms every 10ms of real time, giving us a
	// precision of about 500ms for the mock time.
	mockTime.Resolution = 10 * time.Millisecond
	mockTime.Multiplier = 50
	time.TimeProxy = mockTime
}
