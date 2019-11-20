package graphql

import (
	"github.com/coreos/etcd/etcdserver/etcdserverpb"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
)

var _ schema.ClusterHealthFieldResolvers = (*clusterHealthImpl)(nil)
var _ schema.EtcdAlarmMemberFieldResolvers = (*etcdAlarmMemberImpl)(nil)
var _ schema.EtcdClusterHealthFieldResolvers = (*etcdClusterHealthImpl)(nil)
var _ schema.EtcdClusterMemberHealthFieldResolvers = (*etcdClusterMemberHealthImpl)(nil)

//
// Implement ClusterHealthFieldResolvers
//

type clusterHealthImpl struct{}

// Etcd implements response to request for 'etcd' field.
func (r *clusterHealthImpl) Etcd(p graphql.ResolveParams) (interface{}, error) {
	return p.Source, nil
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
	return string(resp.MemberID), nil
}

//
// Implement EtcdAlarmMemberFieldResolvers
//

type etcdAlarmMemberImpl struct{}

// MemberID implements response to request for 'memberID' field.
func (r *etcdAlarmMemberImpl) MemberID(p graphql.ResolveParams) (string, error) {
	resp := p.Source.(*etcdserverpb.AlarmMember)
	return string(resp.MemberID), nil
}

// MemberID implements response to request for 'memberID' field.
func (r *etcdAlarmMemberImpl) Alarm(p graphql.ResolveParams) (schema.EtcdAlarmType, error) {
	resp := p.Source.(*etcdserverpb.AlarmMember)
	return schema.EtcdAlarmType(resp.Alarm.String()), nil
}
