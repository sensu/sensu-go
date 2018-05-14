package main

import (
	"fmt"
	"github.com/bluele/slack"
)

// Please change these values to suit your environment
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
	msgs, err := api.ChannelsHistoryMessages(&slack.ChannelsHistoryOpt{
		Channel: channel.Id,
	})
	if err != nil {
		panic(err)
	}
	for _, msg := range msgs {
		fmt.Println(msg.UserId, msg.Text)
	}
}
