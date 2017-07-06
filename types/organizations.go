package types

import "errors"

// Define the key type to avoid key collisions in context
type key int

const (
	// OrganizationKey contains the key name to retrieve the org from context
	OrganizationKey key = iota
)

// Organization represents a Sensu organization in RBAC
type Organization struct {
	Description string `json:"description"`
	Name        string `json:"name"`
}

// Validate returns an error if the organization does not pass validation tests
func (o *Organization) Validate() error {
	if err := ValidateName(o.Name); err != nil {
		return errors.New("organization name " + err.Error())
	}

	return nil
}

// FixtureOrganization returns a mocked organization
func FixtureOrganization(org string) *Organization {
	return &Organization{
		Name: org,
	}
}
