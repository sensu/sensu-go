package opampd

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/backend/store"

	"github.com/open-telemetry/opamp-go/protobufs"
)

const (
	entityNamespace = "default"
)

const supportedCapabilities = protobufs.ServerCapabilities_AcceptsStatus | protobufs.ServerCapabilities_OffersRemoteConfig

type Protocol struct {
	Store     store.Store
	PartyMode bool
}

func (p *Protocol) OnStatusReport(instanceUid string, report *protobufs.StatusReport) (*protobufs.ServerToAgent, *corev2.Event, error) {
	entity, err := p.createEntity(instanceUid, report)
	if err != nil {
		return nil, nil, err
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
			atomic.AddUint64(&errorCount, 1)
			return s2a, nil, nil
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
		atomic.AddUint64(&agentConfigsSent, 1)
	}

	var event *corev2.Event
	if report.RemoteConfigStatus != nil {
		event = corev2.NewEvent(entity.ObjectMeta)
		event.Name = ""
		event.Entity = entity
		event.Check = corev2.NewCheck(&corev2.CheckConfig{})
		event.Timestamp = time.Now().Unix()
		event.Check.Name = entity.Name
		event.Check.Namespace = entity.Namespace
		event.Check.Handlers = []string{"dummy"}

		switch report.RemoteConfigStatus.Status {
		case protobufs.RemoteConfigStatus_Failed:
			event.Check.Status = 1
			event.Check.State = corev2.EventFailingState
			event.Check.Output = report.RemoteConfigStatus.ErrorMessage

		case protobufs.RemoteConfigStatus_Applying:
			// status code 3 meant to represent "applying"
			event.Check.Status = 3
			event.Check.State = "applying"

		case protobufs.RemoteConfigStatus_Applied:
			event.Check.Status = 0
			event.Check.State = corev2.EventPassingState
			event.Check.LastOK = time.Now().Unix()
		}
	} else {
		logger.Infof("no RemoteConfigStatus in StatusReport, a demo event will be generated")

		event = corev2.NewEvent(entity.ObjectMeta)
		event.Name = ""
		event.Entity = entity
		event.Check = corev2.NewCheck(&corev2.CheckConfig{})
		event.Timestamp = time.Now().Unix()

		event.Check.Name = entity.Name
		event.Check.Namespace = entity.Namespace
		event.Check.Handlers = []string{"dummy"}
		event.Check.Status = 0
		event.Check.State = corev2.EventPassingState
		event.Check.LastOK = time.Now().Unix()
	}

	return s2a, event, nil
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

func (p *Protocol) createEntity(instanceUid string, report *protobufs.StatusReport) (*corev2.Entity, error) {
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

	return entity, p.Store.UpdateEntity(context.WithValue(context.Background(), corev2.NamespaceKey, entityNamespace), entity)
}
