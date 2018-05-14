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
	channel, err := api.FindChannelByName(channelName)
	if err != nil {
		panic(err)
	}
	err = api.ChatPostMessage(channel.Id, "Hello, world!", &slack.ChatPostMessageOpt{
		Attachments: []*slack.Attachment{
			{Text: "danger", Color: "danger"},
		},
	})
	if err != nil {
		panic(err)
	}
}
