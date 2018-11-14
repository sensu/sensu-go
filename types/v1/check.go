package v1

import (
	"github.com/sensu/sensu-go/types/dynamic"
)

// CheckResult contains the 1.x compatible check result payload
type CheckResult struct {
	Client      string   `json:"client"`
	Status      uint32   `json:"status"`
	Command     string   `json:"command"`
	Subscribers []string `json:"subscribers"`
	Interval    uint32   `json:"interval"`
	Name        string   `json:"name"`
	Issued      int64    `json:"issued"`
	Executed    int64    `json:"executed"`
	Duration    float64  `json:"duration"`
	Output      string   `json:"output"`
}

// Get implements govaluate.Parameters
func (c *CheckResult) Get(name string) (interface{}, error) {
	return dynamic.GetField(c, name)
}
