package tessend

import (
	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

const (
	licenseStorePath = "/sensu.io/api/enterprise/licensing/v2/license"
)

var (
	// Distribution indicates that Sensu Go was distributed from OSS but can be
	// overridden
	Distribution = "oss"
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

	// Distribution indicates how Sensu Go was distributed
	Distribution string `json:"distribution"`

	// Version is the Sensu release version (e.g. "1.4.1").
	Version string `json:"version"`

	// License is the Cluster's license.
	License interface{} `json:"license"`
}

// LicenseFile represents the content of a license file
type LicenseFile struct {
	// License contains the actual license
	License interface{} `json:"license"`
}

// Wrapper wraps the LicenseFile for unmarshalling
type Wrapper struct {
	Value LicenseFile `json:"spec"`
}

// StoreConfig contains information about what services Sensu is using to store
// information.
type StoreConfig struct {
	// ConfigStore indicates the storage backend used to store Sensu configuration.
	ConfigStore string `json:"config_store"`

	// StateStore indicates the storage backend used to store Sensu state.
	StateStore string `json:"state_store"`

	// EventStore indicates the storage backend used to store Sensu events.
	EventStore string `json:"event_store"`
}
