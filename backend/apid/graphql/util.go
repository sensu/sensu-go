package graphql

import (
	"sort"
	"time"

	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/types"
)

// clampInt returns int within given range.
func clampInt(num, min, max int) int {
	if num <= min {
		return min
	} else if num >= max {
		return max
	}
	return num
}

// maxUint32 returns larger of x or y.
func maxUint32(x, y uint32) uint32 {
	if x > y {
		return x
	}
	return y
}

// maxUint32 returns larger of x or y.
func clampSlice(low, high, len int) (int, int) {
	low = clampInt(low, 0, len)
	high = clampInt(high, low, len)
	return low, high
}

// convertTimestamp to instance of time.Time
func convertTs(ts int64) *time.Time {
	if ts == 0 {
		return nil
	}
	t := time.Unix(ts, 0)
	return &t
}

// sortEvents by given enum value
func sortEvents(evs []*types.Event, order schema.EventsListOrder) {
	if order == schema.EventsListOrders.SEVERITY {
		sort.Sort(types.EventsBySeverity(evs))
	} else if order == schema.EventsListOrders.LASTOK {
		sort.Sort(types.EventsByLastOk(evs))
	} else {
		sort.Sort(types.EventsByTimestamp(
			evs,
			order == schema.EventsListOrders.NEWEST,
		))
	}
}
