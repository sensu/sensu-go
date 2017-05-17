package types

// User describes an authenticated user
type User struct {
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
}

// FixtureEntity returns a testing fixture for an Entity object.
func FixtureUser(username string) *User {
	return &User{
		Username: username,
	}
}
