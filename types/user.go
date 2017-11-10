package types

import (
	"errors"
	fmt "fmt"
)

// FixtureUser returns a testing fixture for an Entity object.
func FixtureUser(username string) *User {
	return &User{
		Username: username,
		Password: "P@ssw0rd!",
		Roles:    []string{"default"},
	}
}

// Validate returns an error if the entity is invalid.
func (u *User) Validate() error {
	if err := ValidateNameStrict(u.Username); err != nil {
		return fmt.Errorf("username %s", err)
	}

	return nil
}

// ValidatePassword returns an error if the entity is invalid.
func (u *User) ValidatePassword() error {
	if u.Password == "" {
		return errors.New("password can't be empty")
	}

	if len(u.Password) < 8 {
		return errors.New("password length must be at least 8 characters")
	}

	return nil
}
