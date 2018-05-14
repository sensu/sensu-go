package main

import (
	"github.com/bluele/slack"
)

const (
	token       = "your-api-token"
	channelName = "general"
)

func main() {
	api := slack.New(token)
	err := api.ChatPostMessage(channelName, "Hello, world!", nil)
	if err != nil {
		panic(err)
	}
}
