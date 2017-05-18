package types

import "errors"

// User describes an authenticated user
type User struct {
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
}

// FixtureUser returns a testing fixture for an Entity object.
func FixtureUser(username string) *User {
	return &User{
		Username: username,
	}
}

// Validate returns an error if the entity is invalid.
func (u *User) Validate() error {
	if u.Username == "" {
		return errors.New("username can't be empty")
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
