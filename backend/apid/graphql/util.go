package graphql

import (
	"sort"
	"time"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
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

// minInt returns smaller of x or y.
func minInt(x, y int) int {
	if x > y {
		return y
	}
	return x
}

// maxInt returns larger of x or y.
func maxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}

// clampSlice ensures given indexes are within bounds.
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
func sortEvents(evs []*corev2.Event, order schema.EventsListOrder) {
	if order == schema.EventsListOrders.SEVERITY {
		sort.Sort(corev2.EventsBySeverity(evs, false))
	} else if order == schema.EventsListOrders.LASTOK {
		sort.Sort(corev2.EventsByLastOk(evs, false))
	} else if order == schema.EventsListOrders.ENTITY || order == schema.EventsListOrders.ENTITY_DESC {
		sort.Sort(corev2.EventsByEntityName(evs, order == schema.EventsListOrders.ENTITY))
	} else {
		sort.Sort(corev2.EventsByTimestamp(
			evs,
			order == schema.EventsListOrders.NEWEST,
		))
	}
}
