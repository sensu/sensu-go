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
	auth, err := api.AuthTest()
	if err != nil {
		panic(err)
	}
	fmt.Println(auth.Url)
	fmt.Println(auth.Team)
	fmt.Println(auth.User)
}
