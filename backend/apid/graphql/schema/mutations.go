package schema

// GetOrganization returns organization
func (ns *NamespaceInput) GetOrganization() string {
	return ns.Organization
}

// GetEnvironment returns environment
func (ns *NamespaceInput) GetEnvironment() string {
	return ns.Environment
}
