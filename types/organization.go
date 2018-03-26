package types

import (
	fmt "fmt"
	"net/url"
)

// Validate returns an error if the organization does not pass validation tests
func (o *Organization) Validate() error {
	if err := ValidateName(o.Name); err != nil {
		return fmt.Errorf("organization name %s", err)
	}

	return nil
}

// FixtureOrganization returns a mocked organization
func FixtureOrganization(name string) *Organization {
	return &Organization{
		Name: name,
	}
}

// URIPath returns the path component of a Organization URI.
func (o *Organization) URIPath() string {
	return fmt.Sprintf("/rbac/organizations/%s", url.PathEscape(o.Name))
}
