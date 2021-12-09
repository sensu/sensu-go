package opampd

import (
	"context"
	"fmt"

	"github.com/open-telemetry/opamp-go/protobufs"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"
)

const (
	entityNamespace = "default"
)

const supportedCapabilities = protobufs.ServerCapabilities_AcceptsStatus | protobufs.ServerCapabilities_OffersRemoteConfig

type Protocol struct {
	Store     store.Store
	PartyMode bool
}

func (p *Protocol) OnStatusReport(instanceUid string, report *protobufs.StatusReport) (*protobufs.ServerToAgent, error) {
	err := p.createEntity(instanceUid, report)
	if err != nil {
		return nil, err
	}

	c := report.Capabilities
	s2a := &protobufs.ServerToAgent{
		InstanceUid:  instanceUid,
		Capabilities: supportedCapabilities,
	}
	if (c&protobufs.AgentCapabilities_AcceptsRemoteConfig) > 0 || p.PartyMode {
		logger.Infof("updating remote agent config for agent %s", instanceUid)
		cfg, err := p.Store.GetAgentConfig(context.Background())
		if err != nil {
			logger.WithError(err).Warn("unable to provide opamp agent with remote configuration")
			return s2a, nil
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

func (p *Protocol) createEntity(instanceUid string, report *protobufs.StatusReport) error {
	agentType := ""
	agentVersion := ""
	if report.AgentDescription != nil {
		agentType = report.AgentDescription.AgentType
		agentVersion = report.AgentDescription.AgentVersion
	}
	entity := &corev2.Entity{
		EntityClass:        corev2.EntityOpAMPClass,
		System:             corev2.System{},
		Subscriptions:      []string{"opamp"},
		LastSeen:           0,
		Deregister:         true,
		Deregistration:     corev2.Deregistration{},
		User:               "",
		ExtendedAttributes: nil,
		Redact:             nil,
		ObjectMeta: corev2.ObjectMeta{
			Name:      instanceUid,
			Namespace: entityNamespace,
			Labels: map[string]string{
				"opamp-instanceid":    instanceUid,
				"opamp-agent-type":    agentType,
				"opamp-agent-version": agentVersion,
			},
			Annotations: nil,
			CreatedBy:   "",
		},
		SensuAgentVersion: "",
		KeepaliveHandlers: nil,
	}

	return p.Store.UpdateEntity(context.WithValue(context.Background(), corev2.NamespaceKey, entityNamespace), entity)
}
