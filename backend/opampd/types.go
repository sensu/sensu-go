package opampd

import (
	"github.com/open-telemetry/opamp-go/protobufs"
)

type MessageHandler interface {
	OnStatusReport(instanceUid string, report *protobufs.StatusReport) (*protobufs.ServerToAgent, error)

	OnAddonStatuses(instanceUid string, status *protobufs.AgentAddonStatuses) (*protobufs.ServerToAgent, error)

	OnAgentInstallStatus(instanceUid string, status *protobufs.AgentInstallStatus) (*protobufs.ServerToAgent, error)

	OnAgentDisconnect(instanceUid string, disconnect *protobufs.AgentDisconnect) (*protobufs.ServerToAgent, error)
}
