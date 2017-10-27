package types

import (
	"errors"
)

// Validate returns an error if the entity is invalid.
func (e *Entity) Validate() error {
	if err := ValidateName(e.ID); err != nil {
		return errors.New("entity id " + err.Error())
	}

	if err := ValidateName(e.Class); err != nil {
		return errors.New("entity class " + err.Error())
	}

	if e.Environment == "" {
		return errors.New("environment must be set")
	}

	if e.Organization == "" {
		return errors.New("organization must be set")
	}

	return nil
}

// GetOrg refers to the organization the check belongs to
func (e *Entity) GetOrg() string {
	return e.Organization
}

// GetEnv refers to the organization the check belongs to
func (e *Entity) GetEnv() string {
	return e.Environment
}

// FixtureEntity returns a testing fixture for an Entity object.
func FixtureEntity(id string) *Entity {
	return &Entity{
		ID:               id,
		Class:            "host",
		Subscriptions:    []string{"subscription"},
		Environment:      "default",
		Organization:     "default",
		KeepaliveTimeout: 120,
	}
}
