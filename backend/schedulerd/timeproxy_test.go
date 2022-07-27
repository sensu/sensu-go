package schedulerd

import (
	"github.com/echlebek/crock"
	time "github.com/echlebek/timeproxy"
)

var mockTime = crock.NewTime(time.Unix(42, 0))

func init() {
	mockTime.Resolution = time.Millisecond
	mockTime.Multiplier = 100
	time.TimeProxy = mockTime
}
