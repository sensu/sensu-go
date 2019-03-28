package tessend

import (
	"encoding/json"
	"fmt"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

const (
	// TimestampFormat is the string representation for the license timestamps e.g.
	// "2018-07-26T12:12:06-04:00"
	TimestampFormat = time.RFC3339

	licenseStorePath = "/sensu.io/api/enterprise/licensing/v2/license"
)

// Data is the payload sent to tessen
type Data struct {
	// General information about the Sensu installation.
	Cluster Cluster `json:"cluster"`

	// Metric data.
	Metrics corev2.Metrics `json:"metrics"`
}

// Cluster is the cluster information sent to tessen
type Cluster struct {
	// ID is the ID of the sensu-enterprise-go cluster.
	ID string `json:"id"`

	// Version is the Sensu release version (e.g. "1.4.1").
	Version string `json:"sensu_version"`

	// License is the Cluster's license.
	License License
}

// Timestamp is an alias to time.Time with json Marshaling/Unmarshaling support
type Timestamp time.Time

// MarshalJSON implements the json.Marshaler interface.
func (t Timestamp) MarshalJSON() ([]byte, error) {
	stamp := fmt.Sprintf("%q", time.Time(t).Format(TimestampFormat))
	return []byte(stamp), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface.
func (t *Timestamp) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return fmt.Errorf("Cannot unmarshal the license timestamp")
	}

	value, err := time.Parse(TimestampFormat, str)
	if err != nil {
		return err
	}

	*t = Timestamp(value)
	return nil
}

// License holds information about a user's enterprise software license,
// including duration of validity and enabled features.
type License struct {
	// Version is the license format version.
	Version int `json:"version"`

	// Issuer is the name of the account that issued the license.
	Issuer string `json:"issuer"`

	// AccountName is the name of the customer account.
	AccountName string `json:"accountName"`

	// AccountID is the ID of the customer account.
	AccountID uint64 `json:"accountID"`

	// Issued is the time at which the license was issued.
	Issued Timestamp `json:"issued"`

	// ValidUntil is the time at which the license will expire.
	ValidUntil Timestamp `json:"validUntil"`

	// Plan is the subscription plan the license is associated with.
	Plan string `json:"plan"`

	// Features are a list of features enabled by this license.
	Features []string `json:"features"`
}

// LicenseFile represents the content of a license file
type LicenseFile struct {
	// License contains the actual license
	License License `json:"license"`
}

// Wrapper wraps the LicenseFile for unmarshalling
type Wrapper struct {
	Value LicenseFile `json:"spec"`
}
