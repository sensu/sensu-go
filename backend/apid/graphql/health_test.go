package graphql

import (
	"reflect"
	"testing"

	"github.com/coreos/etcd/etcdserver/etcdserverpb"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/apid/graphql/schema"
	"github.com/sensu/sensu-go/graphql"
)

func Test_clusterHealthImpl_Etcd(t *testing.T) {
	tests := []struct {
		name    string
		source  interface{}
		want    interface{}
		wantErr bool
	}{
		{
			name:    "main",
			source:  "fred",
			want:    "fred",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &clusterHealthImpl{}
			got, err := r.Etcd(graphql.ResolveParams{Source: tt.source})
			if (err != nil) != tt.wantErr {
				t.Errorf("clusterHealthImpl.Etcd() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("clusterHealthImpl.Etcd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_etcdClusterHealthImpl_Alarms(t *testing.T) {
	defaultResp := corev2.FixtureHealthResponse(true)
	tests := []struct {
		name    string
		source  *corev2.HealthResponse
		want    interface{}
		wantErr bool
	}{
		{
			name:    "main",
			source:  defaultResp,
			want:    defaultResp.Alarms,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &etcdClusterHealthImpl{}
			got, err := r.Alarms(graphql.ResolveParams{Source: tt.source})
			if (err != nil) != tt.wantErr {
				t.Errorf("etcdClusterHealthImpl.Alarms() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("etcdClusterHealthImpl.Alarms() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_etcdClusterHealthImpl_Members(t *testing.T) {
	defaultResp := corev2.FixtureHealthResponse(true)
	tests := []struct {
		name    string
		source  *corev2.HealthResponse
		want    interface{}
		wantErr bool
	}{
		{
			name:    "main",
			source:  defaultResp,
			want:    defaultResp.ClusterHealth,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &etcdClusterHealthImpl{}
			got, err := r.Members(graphql.ResolveParams{Source: tt.source})
			if (err != nil) != tt.wantErr {
				t.Errorf("etcdClusterHealthImpl.Members() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("etcdClusterHealthImpl.Members() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_etcdAlarmMemberImpl_MemberID(t *testing.T) {
	defaultResp := corev2.FixtureHealthResponse(true)
	tests := []struct {
		name    string
		source  *etcdserverpb.AlarmMember
		want    string
		wantErr bool
	}{
		{
			name:    "main",
			source:  defaultResp.Alarms[0],
			want:    string(defaultResp.Alarms[0].MemberID),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &etcdAlarmMemberImpl{}
			got, err := r.MemberID(graphql.ResolveParams{Source: tt.source})
			if (err != nil) != tt.wantErr {
				t.Errorf("etcdAlarmMemberImpl.MemberID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("etcdAlarmMemberImpl.MemberID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_etcdAlarmMemberImpl_Alarm(t *testing.T) {
	defaultResp := corev2.FixtureHealthResponse(true)
	tests := []struct {
		name    string
		source  *etcdserverpb.AlarmMember
		want    schema.EtcdAlarmType
		wantErr bool
	}{
		{
			name:    "main",
			source:  defaultResp.Alarms[0],
			want:    schema.EtcdAlarmType(defaultResp.Alarms[0].Alarm.String()),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &etcdAlarmMemberImpl{}
			got, err := r.Alarm(graphql.ResolveParams{Source: tt.source})
			if (err != nil) != tt.wantErr {
				t.Errorf("etcdAlarmMemberImpl.Alarm() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("etcdAlarmMemberImpl.Alarm() = %#v, want [%#v]", got, tt.want)
			}
		})
	}
}

func Test_etcdClusterMemberHealthImpl_MemberID(t *testing.T) {
	defaultResp := corev2.FixtureHealthResponse(true)
	tests := []struct {
		name    string
		source  *corev2.ClusterHealth
		want    string
		wantErr bool
	}{
		{
			name:    "main",
			source:  defaultResp.ClusterHealth[0],
			want:    string(defaultResp.ClusterHealth[0].MemberID),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &etcdClusterMemberHealthImpl{}
			got, err := r.MemberID(graphql.ResolveParams{Source: tt.source})
			if (err != nil) != tt.wantErr {
				t.Errorf("etcdClusterMemberHealthImpl.MemberID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("etcdClusterMemberHealthImpl.MemberID() = %v, want %v", got, tt.want)
			}
		})
	}
}
