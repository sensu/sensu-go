package slack

import (
	"encoding/json"
	"errors"
)

// API groups.list: Lists private groups that the calling user has access to.
func (sl *Slack) GroupsList() ([]*Group, error) {
	uv := sl.urlValues()
	body, err := sl.GetRequest(groupsListApiEndpoint, uv)
	if err != nil {
		return nil, err
	}
	res := new(GroupsListAPIResponse)
	err = json.Unmarshal(body, res)
	if err != nil {
		return nil, err
	}
	if !res.Ok {
		return nil, errors.New(res.Error)
	}
	return res.Groups()
}

// API groups.create: Creates a private group.
func (sl *Slack) CreateGroup(name string) error {
	uv := sl.urlValues()
	uv.Add("name", name)

	_, err := sl.GetRequest(groupsCreateApiEndpoint, uv)
	if err != nil {
		return err
	}
	return nil
}

// slack group type
type Group struct {
	Id         string          `json:"id"`
	Name       string          `json:"name"`
	Created    int             `json:"created"`
	Creator    string          `json:"creator"`
	IsArchived bool            `json:"is_archived"`
	Members    []string        `json:"members"`
	RawTopic   json.RawMessage `json:"topic"`
	RawPurpose json.RawMessage `json:"purpose"`
}

// response type for `groups.list` api
type GroupsListAPIResponse struct {
	BaseAPIResponse
	RawGroups json.RawMessage `json:"groups"`
}

// Groups returns a slice of group object from `groups.list` api.
func (res *GroupsListAPIResponse) Groups() ([]*Group, error) {
	var groups []*Group
	err := json.Unmarshal(res.RawGroups, &groups)
	if err != nil {
		return nil, err
	}
	return groups, nil
}

// response type for `groups.create` api
type GroupsCreateAPIResponse struct {
	BaseAPIResponse
	RawGroup json.RawMessage `json:"group"`
}

func (res *GroupsCreateAPIResponse) Group() (*Group, error) {
	group := Group{}
	err := json.Unmarshal(res.RawGroup, &group)
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// FindGroup returns a group object that satisfy conditions specified.
func (sl *Slack) FindGroup(cb func(*Group) bool) (*Group, error) {
	groups, err := sl.GroupsList()
	if err != nil {
		return nil, err
	}
	for _, group := range groups {
		if cb(group) {
			return group, nil
		}
	}
	return nil, errors.New("No such group.")
}

// FindGroupByName returns a group object that matches name specified.
func (sl *Slack) FindGroupByName(name string) (*Group, error) {
	return sl.FindGroup(func(group *Group) bool {
		return group.Name == name
	})
}

// API groups.invite: Invites a user to a private group.
func (sl *Slack) InviteGroup(channelId, userId string) error {
	uv := sl.urlValues()
	uv.Add("channel", channelId)
	uv.Add("user", userId)

	_, err := sl.GetRequest(channelsJoinApiEndpoint, uv)
	if err != nil {
		return err
	}
	return nil
}
