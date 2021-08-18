package v2

import (
	"errors"
	"fmt"
)

// validate checks if a resource reference resource passes validation rules.
func (r *ResourceReference) Validate() error {
	if err := ValidateName(r.Name); err != nil {
		return errors.New("name " + err.Error())
	}

	if r.Type == "" {
		return errors.New("type must be set")
	}

	if r.APIVersion == "" {
		return errors.New("api_version must be set")
	}

	return nil
}

func (r *ResourceReference) ResourceID() string {
	return fmt.Sprintf("%s.%s(Name=%s)", r.APIVersion, r.Type, r.Name)
}
