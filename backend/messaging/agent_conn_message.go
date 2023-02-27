package messaging

// AgentNotififcation is used to notify subscribers about the connection
// state of an agent entity, identified by its namespace and name.
type AgentNotification struct {
	Namespace string
	Name      string
	Connected bool
}
