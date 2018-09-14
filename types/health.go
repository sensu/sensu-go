package types

import "github.com/coreos/etcd/etcdserver/etcdserverpb"

// ClusterHealth holds cluster member status info.
type ClusterHealth struct {
	// MemberID is the etcd cluster member's ID.
	MemberID uint64
	// Name is the cluster member's name.
	Name string
	// Err holds the string representation of any errors encountered while checking the member's health.
	Err string
	// Healthy describes the health of the cluster member.
	Healthy bool
}

// HealthResponse contains cluster health and cluster alarms.
type HealthResponse struct {
	// Alarms is the list of active etcd alarms.
	Alarms []*etcdserverpb.AlarmMember
	// ClusterHealth is the list of health status for every cluster member.
	ClusterHealth []*ClusterHealth
}
