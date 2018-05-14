package main

import (
	"github.com/bluele/slack"
)

// Please change these values to suit your environment
const (
	token     = "your-api-token"
	groupName = "group-name"
)

func main() {
	api := slack.New(token)
	err := api.ChatPostMessage(groupName, "Hello, world!", &slack.ChatPostMessageOpt{AsUser: true})
	if err != nil {
		panic(err)
	}
}
