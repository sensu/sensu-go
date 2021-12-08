package opampd

import (
	"context"
	"fmt"

	"github.com/open-telemetry/opamp-go/protobufs"
	"github.com/sensu/sensu-go/backend/store"
)

const supportedCapabilities = protobufs.ServerCapabilities_AcceptsStatus | protobufs.ServerCapabilities_OffersRemoteConfig

type Protocol struct {
	Store     store.OpampStore
	PartyMode bool
}

func (p *Protocol) OnStatusReport(instanceUid string, report *protobufs.StatusReport) (*protobufs.ServerToAgent, error) {
	c := report.Capabilities
	s2a := &protobufs.ServerToAgent{
		InstanceUid:  instanceUid,
		Capabilities: supportedCapabilities,
	}
	if (c&protobufs.AgentCapabilities_AcceptsRemoteConfig) > 0 || p.PartyMode {
		logger.Infof("updating remote agent config for agent %s", instanceUid)
		cfg, err := p.Store.GetAgentConfig(context.Background())
		if err != nil {
			return s2a, err
		}
		s2a.RemoteConfig = &protobufs.AgentRemoteConfig{
			Config: &protobufs.AgentConfigMap{
				ConfigMap: map[string]*protobufs.AgentConfigFile{
					"sensu.io": {
						Body:        []byte(cfg.Body),
						ContentType: cfg.ContentType,
					},
				},
			},
		}
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
