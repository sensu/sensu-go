package v1

import (
	"github.com/sensu/sensu-go/types/dynamic"
)

// CheckResult contains the 1.x compatible check result payload
type CheckResult struct {
	Client             string   `json:"client"`
	Status             int32    `json:"status"`
	Command            string   `json:"command"`
	Subscribers        []string `json:"subscribers"`
	Interval           uint32   `json:"interval"`
	Name               string   `json:"name"`
	Issued             int64    `json:"issued"`
	Executed           int64    `json:"executed"`
	Duration           float64  `json:"duration"`
	Output             string   `json:"output"`
	ExtendedAttributes []byte   `json:"-"`
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (c *CheckResult) UnmarshalJSON(b []byte) error {
	return dynamic.Unmarshal(b, c)
}

// MarshalJSON implements the json.Marshaler interface.
func (c *CheckResult) MarshalJSON() ([]byte, error) {
	return dynamic.Marshal(c)
}

// SetExtendedAttributes sets the serialized ExtendedAttributes of c.
func (c *CheckResult) SetExtendedAttributes(e []byte) {
	c.ExtendedAttributes = e
}

// GetExtendedAttributes gets the serialized ExtendedAttributes of c.
func (c *CheckResult) GetExtendedAttributes() []byte {
	if c != nil {
		return c.ExtendedAttributes
	}
	return nil
}

// Get implements govaluate.Parameters
func (c *CheckResult) Get(name string) (interface{}, error) {
	return dynamic.GetField(c, name)
}
