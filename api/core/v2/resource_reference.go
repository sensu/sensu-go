package v2

import (
	"errors"
	"fmt"
)

// Validate checks if a resource reference resource passes validation rules.
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

// ResourceID returns a string that uniquely identifies a ResourceReference
// in the format: APIVersion.Type(Name=%s)
func (r *ResourceReference) ResourceID() string {
	return fmt.Sprintf("%s.%s(Name=%s)", r.APIVersion, r.Type, r.Name)
}

// LogFields returns a map of field names to values which represent a
// ResourceReference.
func (r *ResourceReference) LogFields(debug bool) map[string]interface{} {
	fields := map[string]interface{}{
		"api_version": r.APIVersion,
		"type":        r.Type,
		"name":        r.Name,
	}
	return fields
}
