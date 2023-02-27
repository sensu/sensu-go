package store

// EventGauges are gauge metrics about the events stored in Sensu.
type EventGauges struct {
	Total          int64 `json:"total"`
	StatePassing   int64 `json:"state_passing"`
	StateFailing   int64 `json:"state_failing"`
	StateFlapping  int64 `json:"state_flapping"`
	StatusOK       int64 `json:"status_ok"`
	StatusWarning  int64 `json:"status_warning"`
	StatusCritical int64 `json:"status_critical"`
	StatusOther    int64 `json:"status_other"`
}

type KeepaliveGauges = EventGauges
