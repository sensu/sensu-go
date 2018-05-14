package main

import (
	"fmt"
	"github.com/bluele/slack"
)

const (
	token = "your-api-token"
)

func main() {
	api := slack.New(token)
	channels, err := api.ChannelsList()
	if err != nil {
		panic(err)
	}
	for _, channel := range channels {
		fmt.Println(channel.Id, channel.Name)
	}
}
