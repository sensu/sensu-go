package v2

import (
	"github.com/coreos/etcd/version"
)

// Version holds the current etcd server and cluster version, and the sensu-backend version.
type Version struct {
	Etcd         *version.Versions `json:"etcd"`
	SensuBackend string            `json:"sensu_backend"`
}

// FixtureVersion returns a Version fixture for testing.
func FixtureVersion() *Version {
	version := &Version{
		Etcd: &version.Versions{
			Server:  "3.3.2",
			Cluster: "3.3.0",
		},
		SensuBackend: "5.7.0#20ba7cb",
	}
	return version
}
