package schema

// NewNamespaceInput returns new instance using given values.
func NewNamespaceInput(org, env string) *NamespaceInput {
	return &NamespaceInput{
		Organization: org,
		Environment:  env,
	}
}

// GetOrganization returns organization
func (ns *NamespaceInput) GetOrganization() string {
	return ns.Organization
}

// GetEnvironment returns environment
func (ns *NamespaceInput) GetEnvironment() string {
	return ns.Environment
}
