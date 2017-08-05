package types

import "errors"

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
func FixtureOrganization(name string) *Organization {
	return &Organization{
		Name: name,
	}
}
