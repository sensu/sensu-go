package etcd

import "github.com/prometheus/client_golang/prometheus"

const (
	LeaseOperationsCounterVec = "sensu_go_lease_ops"

	LeaseOperationTypeName   = "op"
	LeaseOperationStatusName = "status"

	LeaseOperationTypeGrant     = "grant"
	LeaseOperationTypeRevoke    = "revoke"
	LeaseOperationTypePut       = "put"
	LeaseOperationTypeKeepalive = "keepalive"
	LeaseOperationTypeFind      = "find"
	LeaseOperationTypeTTL       = "ttl"

	LeaseOperationStatusOK      = "ok"
	LeaseOperationStatusError   = "error"
	LeaseOperationStatusAlive   = "alive"
	LeaseOperationStatusExpired = "expired"
)

var LeaseOperationsCounter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: LeaseOperationsCounterVec,
		Help: "the total number of lease operations",
	},
	[]string{LeaseOperationTypeName, LeaseOperationStatusName},
)

func LeaseStatusFor(err error) string {
	if err == nil {
		return LeaseOperationStatusOK
	}
	return LeaseOperationStatusError
}

func init() {
	if err := prometheus.Register(LeaseOperationsCounter); err != nil {
		panic(err)
	}
}
