package graphql

import (
	"context"
	"strconv"
	"time"

	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
	"go.etcd.io/etcd/api/v3/etcdserverpb"
)

var _ schema.ClusterHealthFieldResolvers = (*clusterHealthImpl)(nil)
var _ schema.EtcdAlarmMemberFieldResolvers = (*etcdAlarmMemberImpl)(nil)
var _ schema.EtcdClusterHealthFieldResolvers = (*etcdClusterHealthImpl)(nil)
var _ schema.EtcdClusterMemberHealthFieldResolvers = (*etcdClusterMemberHealthImpl)(nil)

//
// Implement ClusterHealthFieldResolvers
//

type clusterHealthImpl struct {
	healthController EtcdHealthController
}

// Etcd implements response to request for 'etcd' field.
func (r *clusterHealthImpl) Etcd(p schema.ClusterHealthEtcdFieldResolverParams) (interface{}, error) {
	ctx, cancel := context.WithTimeout(p.Context, time.Duration(p.Args.Timeout)*time.Millisecond)
	defer cancel()
	resp := r.healthController.GetClusterHealth(ctx)
	return resp, nil
}

//
// Implement EtcdClusterHealthFieldResolvers
//

type etcdClusterHealthImpl struct{}

// Alarms implements response to request for 'alarms' field.
func (r *etcdClusterHealthImpl) Alarms(p graphql.ResolveParams) (interface{}, error) {
	resp := p.Source.(*corev2.HealthResponse)
	return resp.Alarms, nil
}

// Members implements response to request for 'Members' field.
func (r *etcdClusterHealthImpl) Members(p graphql.ResolveParams) (interface{}, error) {
	resp := p.Source.(*corev2.HealthResponse)
	return resp.ClusterHealth, nil
}

//
// Implement EtcdClusterMemberHealthFieldResolvers
//

type etcdClusterMemberHealthImpl struct {
	schema.EtcdClusterMemberHealthAliases
}

// MemberID implements response to request for 'memberID' field.
func (r *etcdClusterMemberHealthImpl) MemberID(p graphql.ResolveParams) (string, error) {
	resp := p.Source.(*corev2.ClusterHealth)
	return strconv.FormatUint(resp.MemberID, 10), nil
}

//
// Implement EtcdAlarmMemberFieldResolvers
//

type etcdAlarmMemberImpl struct{}

// MemberID implements response to request for 'memberID' field.
func (r *etcdAlarmMemberImpl) MemberID(p graphql.ResolveParams) (string, error) {
	resp := p.Source.(*etcdserverpb.AlarmMember)
	return strconv.FormatUint(resp.MemberID, 10), nil
}

// MemberID implements response to request for 'memberID' field.
func (r *etcdAlarmMemberImpl) Alarm(p graphql.ResolveParams) (schema.EtcdAlarmType, error) {
	resp := p.Source.(*etcdserverpb.AlarmMember)
	return schema.EtcdAlarmType(resp.Alarm.String()), nil
}
