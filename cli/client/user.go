package client

import (
	"encoding/json"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/types"
)

var usersPath = CreateBasePath(coreAPIGroup, coreAPIVersion, "users")

// AddGroupToUser makes "username" a member of "group".
func (client *RestClient) AddGroupToUser(username, group string) error {
	path := usersPath(username, "groups", group)
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
func (client *RestClient) CreateUser(user *types.User) error {
	path := usersPath("")
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
	path := usersPath(username)
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
func (client *RestClient) FetchUser(username string) (*types.User, error) {
	user := &types.User{}
	path := usersPath(username)
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

// ListUsers fetches all users from configured Sensu instance
func (client *RestClient) ListUsers(options *ListOptions) ([]corev2.User, error) {
	var users []corev2.User

	if err := client.List(usersPath(), &users, options); err != nil {
		return users, err
	}

	return users, nil
}

// ReinstateUser reinstates a disabled user on configured Sensu instance
func (client *RestClient) ReinstateUser(username string) error {
	path := usersPath(username, "reinstate")
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
	path := usersPath(username, "groups", group)
	res, err := client.R().Delete(path)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// RemoveGroupsFromUser removes all the groups for "username".
func (client *RestClient) RemoveAllGroupsFromUser(username string) error {
	path := usersPath(username, "groups")
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

// UpdatePassword updates password of given user on configured Sensu instance
func (client *RestClient) UpdatePassword(username, pwd string) error {
	bytes, err := json.Marshal(map[string]string{"password": pwd})
	if err != nil {
		return err
	}

	path := usersPath(username, "password")
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
