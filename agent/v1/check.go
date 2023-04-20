package v1

// CheckResult contains the 1.x compatible check result payload
// This is a subset of 1.x attributes, see:
// https://docs.sensu.io/sensu-core/1.6/reference/checks/#check-attributes
type CheckResult struct {
	Source      string   `json:"source"`
	Status      uint32   `json:"status"`
	Command     string   `json:"command"`
	Subscribers []string `json:"subscribers"`
	Interval    uint32   `json:"interval"`
	Name        string   `json:"name"`
	Issued      int64    `json:"issued"`
	Executed    int64    `json:"executed"`
	Duration    float64  `json:"duration"`
	Output      string   `json:"output"`

	// Client is deprecated but still supported
	Client string `json:"client"`

	// Handler is deprecated but still supported
	Handler string `json:"handler"`

	// Handlers supercedes Handler
	Handlers []string `json:"handlers"`
}
