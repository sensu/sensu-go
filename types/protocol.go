package types

// AgentHandshakeType is the message type string for an AgentHandshake
const AgentHandshakeType = "agent_handshake"

// An AgentHandshake is the first message sent by a Backend on a Transport in a
// Session.
type AgentHandshake Entity

// BackendHandshakeType is the message type string for a BackendHandshake
const BackendHandshakeType = "backend_handshake"

// A BackendHandshake is the first message sent by a Backend on a Transport in
// a Session.
type BackendHandshake struct{}
