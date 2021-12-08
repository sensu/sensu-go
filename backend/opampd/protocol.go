package opampd

import (
	"fmt"

	"github.com/open-telemetry/opamp-go/protobufs"
)

type Protocol struct {
}

func (p *Protocol) OnStatusReport(instanceUid string, report *protobufs.StatusReport) (*protobufs.ServerToAgent, error) {
	//TODO implement me
	s2a := &protobufs.ServerToAgent{
		InstanceUid:           instanceUid,
		ErrorResponse:         nil,
		RemoteConfig:          nil,
		ConnectionSettings:    nil,
		AddonsAvailable:       nil,
		AgentPackageAvailable: nil,
		Flags:                 0,
		Capabilities:          0,
	}
	return s2a, nil
}

func (p *Protocol) OnAddonStatuses(instanceUid string, status *protobufs.AgentAddonStatuses) (*protobufs.ServerToAgent, error) {
	//TODO implement me
	return nil, fmt.Errorf("implement me")
}

func (p *Protocol) OnAgentInstallStatus(instanceUid string, status *protobufs.AgentInstallStatus) (*protobufs.ServerToAgent, error) {
	//TODO implement me
	return nil, fmt.Errorf("implement me")
}

func (p *Protocol) OnAgentDisconnect(instanceUid string, disconnect *protobufs.AgentDisconnect) (*protobufs.ServerToAgent, error) {
	//TODO implement me
	return nil, fmt.Errorf("implement me")
}
