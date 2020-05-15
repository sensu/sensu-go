package client

import (
	"encoding/json"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
)

// UsersPath is the api path for users.
var UsersPath = CreateBasePath(coreAPIGroup, coreAPIVersion, "users")

// AddGroupToUser makes "username" a member of "group".
func (client *RestClient) AddGroupToUser(username, group string) error {
	path := UsersPath(username, "groups", group)
	res, err := client.R().Put(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// CreateUser creates new check on configured Sensu instance
func (client *RestClient) CreateUser(user *corev2.User) error {
	path := UsersPath("")
	res, err := client.R().SetBody(user).Post(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// DisableUser disables a user on configured Sensu instance
func (client *RestClient) DisableUser(username string) error {
	path := UsersPath(username)
	res, err := client.R().Delete(path)

	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// FetchUser retrieve the given user
func (client *RestClient) FetchUser(username string) (*corev2.User, error) {
	user := &corev2.User{}
	path := UsersPath(username)
	res, err := client.R().Get(path)

	if err != nil {
		return nil, err
	}

	if res.StatusCode() >= 400 {
		return nil, UnmarshalError(res)
	}

	err = json.Unmarshal(res.Body(), user)
	return user, err
}

// ReinstateUser reinstates a disabled user on configured Sensu instance
func (client *RestClient) ReinstateUser(username string) error {
	path := UsersPath(username, "reinstate")
	res, err := client.R().Put(path)

	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// RemoveGroupFromUser removes "username" from the given "group".
func (client *RestClient) RemoveGroupFromUser(username, group string) error {
	path := UsersPath(username, "groups", group)
	res, err := client.R().Delete(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// RemoveAllGroupsFromUser removes all the groups for "username".
func (client *RestClient) RemoveAllGroupsFromUser(username string) error {
	path := UsersPath(username, "groups")
	res, err := client.R().Delete(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// SetGroupsForUser sets the groups for "username" to "groups".
func (client *RestClient) SetGroupsForUser(username string, groups []string) error {
	// Note: Instead of the implementation below, we can have the backend
	// support receiving a list of groups on /rbac/users/{username}/groups

	// Start by removing all the existing groups
	if err := client.RemoveAllGroupsFromUser(username); err != nil {
		return err
	}

	// Then add each group in the set one by one
	for _, group := range groups {
		if err := client.AddGroupToUser(username, group); err != nil {
			return err
		}
	}

	return nil
}

// ResetPassword reset the password of given user on configured Sensu instance
func (client *RestClient) ResetPassword(username, passwordHash string) error {
	bytes, err := json.Marshal(map[string]string{
		"password_hash": passwordHash,
	})
	if err != nil {
		return err
	}

	path := UsersPath(username, "reset_password")
	res, err := client.R().
		SetBody(bytes).
		Put(path)

	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// UpdatePassword updates password of given user on configured Sensu instance
func (client *RestClient) UpdatePassword(username, newPasswordHash, currentPassword string) error {
	bytes, err := json.Marshal(map[string]string{
		"password":      currentPassword,
		"password_hash": newPasswordHash,
	})
	if err != nil {
		return err
	}

	path := UsersPath(username, "password")
	res, err := client.R().
		SetBody(bytes).
		Put(path)

	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}
