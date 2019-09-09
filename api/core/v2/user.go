package v2

import (
	"errors"
	fmt "fmt"
	"net/url"
	"path"
	"strconv"
	"strings"
)

const (
	// UsersResource is the name of this resource type
	UsersResource = "users"
)

// GetObjectMeta is a dummy implementation to meet the Resource interface.
func (u *User) GetObjectMeta() ObjectMeta {
	return ObjectMeta{Name: u.Username}
}

// StorePrefix returns the path prefix to this resource in the store
func (u *User) StorePrefix() string {
	return UsersResource
}

// URIPath is the URI path component to a user.
func (u *User) URIPath() string {
	return path.Join(URLPrefix, UsersResource, url.PathEscape(u.Username))
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

// FixtureUser returns a testing fixture for an Entity object.
func FixtureUser(username string) *User {
	return &User{
		Username: username,
		Password: "P@ssw0rd!",
		Groups:   []string{"default"},
	}
}

// UserFields returns a set of fields that represent that resource
func UserFields(r Resource) map[string]string {
	resource := r.(*User)
	return map[string]string{
		"user.username": resource.Username,
		"user.disabled": strconv.FormatBool(resource.Disabled),
		"user.groups":   strings.Join(resource.Groups, ","),
	}
}

// SetNamespace sets the namespace of the resource.
func (u *User) SetNamespace(namespace string) {
}

// SetObjectMeta is a dummy implementation to meet the Resource interface.
func (u *User) SetObjectMeta(meta ObjectMeta) {
	u.Username = meta.Name
}

func (*User) RBACName() string {
	return "users"
}
