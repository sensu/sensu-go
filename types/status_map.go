package types

import "github.com/coreos/etcd/etcdserver/etcdserverpb"

// StatusMap is a map of backend component names to their current status info.
type StatusMap map[string]bool

// Healthy returns true if the StatsMap shows all healthy indicators; false
// otherwise.
func (s StatusMap) Healthy() bool {
	for _, v := range s {
		if !v {
			return false
		}
	}
	return true
}

// ClusterHealth holds cluster member status info.
type ClusterHealth struct {
	// MemberID is the etcd cluster member's ID.
	MemberID uint64
	// Name is the cluster member's name.
	Name string
	// Err holds any errors encountered while checking the member's health.
	Err error
	// Healthy describes the health of the cluster member.
	Healthy bool
}

// HealthResponse contains cluster health and cluster alarms.
type HealthResponse struct {
	Alarms        []*etcdserverpb.AlarmMember
	ClusterHealth []*ClusterHealth
}
