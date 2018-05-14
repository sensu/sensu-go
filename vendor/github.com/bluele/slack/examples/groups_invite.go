package main

import (
	"github.com/bluele/slack"
)

const (
	token          = "your-api-token"
	inviteUserName = "your-team-member"
	groupName      = "your-group-name"
)

func main() {
	api := slack.New(token)
	group, err := api.FindGroupByName(groupName)
	if err != nil {
		panic(err)
	}

	user, err := api.FindUser(func(user *slack.User) bool {
		return user.Name == inviteUserName
	})
	if err != nil {
		panic(err)
	}

	err = api.InviteGroup(group.Id, user.Id)
	if err != nil {
		panic(err)
	}
}
