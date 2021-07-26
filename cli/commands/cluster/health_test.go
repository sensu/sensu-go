package cluster

import (
	"fmt"
	"testing"

	client "github.com/sensu/sensu-go/cli/client/testing"
	test "github.com/sensu/sensu-go/cli/commands/testing"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.etcd.io/etcd/api/v3/etcdserverpb"
)

func TestHealthCommand(t *testing.T) {
	assert := assert.New(t)

	cli := test.NewCLI()
	cmd := HealthCommand(cli)

	assert.NotNil(cmd, "cmd should be returned")
	assert.NotNil(cmd.RunE, "cmd should be able to be executed")
	assert.Regexp("health", cmd.Use)
	assert.Regexp("get sensu health status", cmd.Short)
}

func TestHealthCommandAlarmCorrupt(t *testing.T) {
	assert := assert.New(t)
	clusterID := uint64(4255616304056076734)

	healthResponse := &types.HealthResponse{
		Header: &etcdserverpb.ResponseHeader{
			ClusterId: clusterID,
		},
	}
	clusterHealth := []*types.ClusterHealth{}
	clusterHealth = append(clusterHealth, &types.ClusterHealth{
		MemberID: uint64(12345),
		Name:     "backend0",
		Err:      "",
		Healthy:  true,
	})
	clusterHealth = append(clusterHealth, &types.ClusterHealth{
		MemberID: uint64(12345),
		Name:     "backend1",
		Err:      "error",
		Healthy:  false,
	})

	alarms := []*etcdserverpb.AlarmMember{}
	alarms = append(alarms, &etcdserverpb.AlarmMember{
		MemberID: uint64(56789),
		Alarm:    etcdserverpb.AlarmType_CORRUPT,
	})

	healthResponse.ClusterHealth = clusterHealth
	healthResponse.Alarms = alarms

	cli := test.NewCLIWithValue("none")
	client := cli.Client.(*client.MockClient)
	client.On("Health").Return(healthResponse, nil)

	cmd := HealthCommand(cli)
	require.NoError(t, cmd.Flags().Set("format", "none"))
	out, err := test.RunCmd(cmd, []string{})
	require.NoError(t, err)

	assert.Contains(out, "ID")                         // heading
	assert.Contains(out, "Name")                       // heading
	assert.Contains(out, "Error")                      // heading
	assert.Contains(out, "Healthy")                    // heading
	assert.Contains(out, "Alarm Type")                 // Heading
	assert.Contains(out, "true")                       // healthy cluster member
	assert.Contains(out, "false")                      // unhealthy cluster member
	assert.Contains(out, "error")                      // cluster error
	assert.Contains(out, "CORRUPT")                    // alarm type
	assert.Contains(out, "Cluster ID")                 // cluster id title
	assert.Contains(out, fmt.Sprintf("%x", clusterID)) // cluster id hex
}

func TestHealthCommandAlarmNoSpace(t *testing.T) {
	assert := assert.New(t)
	clusterID := uint64(4255616304056076734)

	healthResponse := &types.HealthResponse{
		Header: &etcdserverpb.ResponseHeader{
			ClusterId: clusterID,
		},
	}
	clusterHealth := []*types.ClusterHealth{}
	clusterHealth = append(clusterHealth, &types.ClusterHealth{
		MemberID: uint64(12345),
		Name:     "backend1",
		Err:      "error",
		Healthy:  false,
	})

	alarms := []*etcdserverpb.AlarmMember{}
	alarms = append(alarms, &etcdserverpb.AlarmMember{
		MemberID: uint64(56789),
		Alarm:    etcdserverpb.AlarmType_NOSPACE,
	})

	healthResponse.ClusterHealth = clusterHealth
	healthResponse.Alarms = alarms

	cli := test.NewCLIWithValue("none")
	client := cli.Client.(*client.MockClient)
	client.On("Health").Return(healthResponse, nil)

	cmd := HealthCommand(cli)
	require.NoError(t, cmd.Flags().Set("format", "none"))
	out, err := test.RunCmd(cmd, []string{})
	require.NoError(t, err)

	assert.Contains(out, "ID")                         // heading
	assert.Contains(out, "Name")                       // heading
	assert.Contains(out, "Error")                      // heading
	assert.Contains(out, "Healthy")                    // heading
	assert.Contains(out, "Alarm Type")                 // Heading
	assert.Contains(out, "false")                      // unhealthy cluster member
	assert.Contains(out, "error")                      // cluster error
	assert.Contains(out, "NOSPACE")                    // alarm type
	assert.Contains(out, "Cluster ID")                 // cluster id title
	assert.Contains(out, fmt.Sprintf("%x", clusterID)) // cluster id hex
}
