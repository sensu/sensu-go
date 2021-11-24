package v2

import (
	"sort"
)

const (
	EventSortEntity    = "ENTITY"
	EventSortLastOk    = "LASTOK"
	EventSortSeverity  = "SEVERITY"
	EventSortTimestamp = "TIMESTAMP"
)

func SortEvents(events []*Event, ordering string, descending bool) {
	if len(ordering) == 0 {
		return
	}

	asc := !descending

	var sortIf sort.Interface
	switch ordering {
	case EventSortEntity:
		sortIf = EventsByEntityName(events, asc)
	case EventSortLastOk:
		sortIf = EventsByLastOk(events, asc)
	case EventSortSeverity:
		sortIf = EventsBySeverity(events, asc)
	default:
		sortIf = EventsByTimestamp(events, asc)
	}

	sort.Sort(sortIf)
}

// EventsBySeverity can be used to sort a given collection of events by check
// status and timestamp.
func EventsBySeverity(es []*Event, asc bool) sort.Interface {
	return &eventSorter{es, createCmpEvents(
		cmpBySeverity(asc),
		cmpByLastOk(asc),
		cmpByEntity(true),
	)}
}

// EventsByTimestamp can be used to sort a given collection of events by time it
// occurred.
func EventsByTimestamp(es []*Event, asc bool) sort.Interface {
	sorter := &eventSorter{events: es}
	if asc {
		sorter.byFn = func(a, b *Event) bool {
			return a.Timestamp > b.Timestamp
		}
	} else {
		sorter.byFn = func(a, b *Event) bool {
			return a.Timestamp < b.Timestamp
		}
	}
	return sorter
}

// EventsByLastOk can be used to sort a given collection of events by time it
// last received an OK status.
func EventsByLastOk(es []*Event, asc bool) sort.Interface {
	return &eventSorter{es, createCmpEvents(
		cmpByState(asc),
		cmpByLastOk(asc),
		cmpByEntity(true),
	)}
}

// EventsByEntityName can be used to sort a given collection of events by
// entity name (and secondarily by check name.)
func EventsByEntityName(es []*Event, asc bool) sort.Interface {
	return &eventSorter{es, createCmpEvents(cmpByEntity(asc))}
}

func cmpByEntity(asc bool) func(a, b *Event) int {
	return func(a, b *Event) int {
		ai, bi := "", ""
		if a.Entity != nil {
			ai += a.Entity.Name
		}
		if a.Check != nil {
			ai += a.Check.Name
		}
		if b.Entity != nil {
			bi = b.Entity.Name
		}
		if b.Check != nil {
			bi += b.Check.Name
		}

		if ai == bi {
			return 0
		} else if (asc && ai < bi) || (!asc && ai > bi) {
			return 1
		}
		return -1
	}
}

func cmpBySeverity(asc bool) func(a, b *Event) int {
	return func(a, b *Event) int {
		ap, bp := deriveSeverity(a), deriveSeverity(b)

		// Sort events with the same exit status by timestamp
		if ap == bp {
			return 0
		}
		if (asc && ap > bp) || (!asc && ap < bp) {
			return 1
		}
		return -1
	}
}

func cmpByState(asc bool) func(a, b *Event) int {
	return func(a, b *Event) int {
		var av, bv bool
		if a.Check != nil {
			av = a.Check.State == EventPassingState
		}
		if b.Check != nil {
			bv = b.Check.State == EventPassingState
		}

		// Rank higher if failing/flapping
		if av == bv {
			return 0
		}
		if (asc && av) || (!asc && !av) {
			return 1
		}
		return -1
	}
}

func cmpByLastOk(asc bool) func(a, b *Event) int {
	return func(a, b *Event) int {
		at, bt := a.Timestamp, b.Timestamp
		if a.HasCheck() {
			at = a.Check.LastOK
		}
		if b.HasCheck() {
			bt = b.Check.LastOK
		}

		if at == bt {
			return 0
		}
		if (asc && at < bt) || (!asc && at > bt) {
			return 1
		}
		return -1
	}
}

// Based on convention we define the order of importance as critical (2),
// warning (1), unknown (>2), and Ok (0). If event is not a check sort to
// very end.
func deriveSeverity(e *Event) int {
	if e.HasCheck() {
		switch e.Check.Status {
		case 0:
			return 3
		case 1:
			return 1
		case 2:
			return 0
		default:
			return 2
		}
	}
	return 4
}

type cmpEvents func(a, b *Event) int

func createCmpEvents(cmps ...cmpEvents) func(a, b *Event) bool {
	return func(a, b *Event) bool {
		for _, cmp := range cmps {
			st := cmp(a, b)
			if st == 0 { // if equal try the next comparitor
				continue
			}
			return st == 1
		}
		return true
	}
}

type eventSorter struct {
	events []*Event
	byFn   func(a, b *Event) bool
}

// Len implements sort.Interface.
func (s *eventSorter) Len() int {
	return len(s.events)
}

// Swap implements sort.Interface.
func (s *eventSorter) Swap(i, j int) {
	s.events[i], s.events[j] = s.events[j], s.events[i]
}

// Less implements sort.Interface.
func (s *eventSorter) Less(i, j int) bool {
	return s.byFn(s.events[i], s.events[j])
}
