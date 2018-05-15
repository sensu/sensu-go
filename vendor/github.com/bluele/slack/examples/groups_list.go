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
	groups, err := api.GroupsList()
	if err != nil {
		panic(err)
	}
	for _, group := range groups {
		fmt.Println(group.Id, group.Name)
	}
}
