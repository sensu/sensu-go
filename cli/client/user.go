package client

import (
	"encoding/json"
	"net/url"

	"github.com/sensu/sensu-go/types"
)

// AddGroupToUser makes "username" a member of "group".
func (client *RestClient) AddGroupToUser(username, group string) error {
	username, group = url.PathEscape(username), url.PathEscape(group)
	res, err := client.R().Put("/rbac/users/" + username + "/groups/" + group)
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// AddRoleToUser adds roles to given user on configured Sensu instance
func (client *RestClient) AddRoleToUser(username, role string) error {
	username, role = url.PathEscape(username), url.PathEscape(role)
	res, err := client.R().Put("/rbac/users/" + username + "/roles/" + role)
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
	res, err := client.R().SetBody(user).Post("/rbac/users")
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
	res, err := client.R().Delete("/rbac/users/" + url.PathEscape(username))

	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// ListUsers fetches all users from configured Sensu instance
func (client *RestClient) ListUsers() ([]types.User, error) {
	var users []types.User

	res, err := client.R().Get("/rbac/users")
	if err != nil {
		return users, err
	}

	if res.StatusCode() >= 400 {
		return users, UnmarshalError(res)
	}

	err = json.Unmarshal(res.Body(), &users)
	return users, err
}

// ReinstateUser reinstates a disabled user on configured Sensu instance
func (client *RestClient) ReinstateUser(uname string) error {
	res, err := client.R().Put("/rbac/users/" + url.PathEscape(uname) + "/reinstate")

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
	username, group = url.PathEscape(username), url.PathEscape(group)
	res, err := client.R().Delete("/rbac/users/" + username + "/groups/" + group)
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
	username = url.PathEscape(username)

	res, err := client.R().Delete("/rbac/users/" + username + "/groups")
	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}

// RemoveRoleFromUser removes role from given user on configured Sensu instance
func (client *RestClient) RemoveRoleFromUser(username, role string) error {
	username, role = url.PathEscape(username), url.PathEscape(role)
	res, err := client.R().Delete("/rbac/users/" + username + "/roles/" + role)
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
	username = url.PathEscape(username)

	// Note: Instead of the implementation below, we can have the backend
	// support receiving a list of groups on /rbac/users/{username}/groups

	// Start by removing all the existing groups
	if err := client.RemoveAllGroupsFromUser(username); err != nil {
		return err
	}

	// Then add each group in the set one by one
	for _, group := range groups {
		if err := client.AddGroupToUser(username, url.PathEscape(group)); err != nil {
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

	res, err := client.R().
		SetBody(bytes).
		Put("/rbac/users/" + url.PathEscape(username) + "/password")

	if err != nil {
		return err
	}

	if res.StatusCode() >= 400 {
		return UnmarshalError(res)
	}

	return nil
}
