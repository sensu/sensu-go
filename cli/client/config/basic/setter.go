package basic

// SetOrganization overrides existing organization value; does not write change
// to file.
func (c *Config) SetOrganization(org string) {
	c.Profile.Organization = org
}

// SetEnvironment overrides existing environment value; does not write change
// to file.
func (c *Config) SetEnvironment(env string) {
	c.Profile.Environment = env
}
