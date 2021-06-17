package v2

import (
	"encoding/json"
	"fmt"

	"go.etcd.io/etcd/api/v3/etcdserverpb"
)

// ClusterHealth holds cluster member status info.
type ClusterHealth struct {
	// MemberID is the etcd cluster member's ID.
	MemberID uint64

	// MemberIDHex is the hexadecimal representation of the member's ID.
	MemberIDHex string

	// Name is the cluster member's name.
	Name string

	// Err contains any error encountered while checking the member's health.
	Err string

	// Healthy describes the health of the cluster member.
	Healthy bool
}

// PostgresHealth holds postgres store status info.
type PostgresHealth struct {
	// Name is the name of the postgres resource.
	Name string

	// Active indicates if the store is configured to use the postgres configuration.
	Active bool

	// Healthy indicates if the postgres store is connected and can query the events table.
	Healthy bool
}

func (h ClusterHealth) MarshalJSON() ([]byte, error) {
	if h.MemberIDHex == "" {
		h.MemberIDHex = fmt.Sprintf("%x", h.MemberID)
	}
	type Clone ClusterHealth
	var clone *Clone = (*Clone)(&h)
	return json.Marshal(clone)
}

// HealthResponse contains cluster health and cluster alarms.
type HealthResponse struct {
	// Alarms is the list of active etcd alarms.
	Alarms []*etcdserverpb.AlarmMember
	// ClusterHealth is the list of health status for every cluster member.
	ClusterHealth []*ClusterHealth
	// Header is the response header for the entire cluster response.
	Header *etcdserverpb.ResponseHeader
	// PostgresHealth is the list of health status for each postgres config.
	PostgresHealth []*PostgresHealth `json:"PostgresHealth,omitempty"`
}

// FixtureHealthResponse returns a HealthResponse fixture for testing.
func FixtureHealthResponse(healthy bool) *HealthResponse {
	var err string
	healthResponse := &HealthResponse{
		Header: &etcdserverpb.ResponseHeader{
			ClusterId: uint64(4255616304056076734),
		},
	}

	clusterHealth := []*ClusterHealth{}
	clusterHealth = append(clusterHealth, &ClusterHealth{
		MemberID: uint64(12345),
		Name:     "backend0",
		Err:      "",
		Healthy:  true,
	})
	if healthy {
		err = ""
	} else {
		err = "cluster error"
	}
	clusterHealth = append(clusterHealth, &ClusterHealth{
		MemberID: uint64(12345),
		Name:     "backend1",
		Err:      err,
		Healthy:  false,
	})

	alarms := []*etcdserverpb.AlarmMember{}
	alarms = append(alarms, &etcdserverpb.AlarmMember{
		MemberID: uint64(56789),
		Alarm:    etcdserverpb.AlarmType_CORRUPT,
	})

	healthResponse.ClusterHealth = clusterHealth
	healthResponse.Alarms = alarms

	return healthResponse
}
