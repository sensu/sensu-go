package v3

import "path"

const OpampAgentConfigResource = "/agentconfig"

func (OpampAgentConfig) uriPath() string {
	return path.Join("/api/opamp", OpampAgentConfigResource)
}
