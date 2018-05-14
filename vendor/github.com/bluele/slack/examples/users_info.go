package main

import (
	"fmt"
	"github.com/bluele/slack"
)

// Please change these values to suit your environment
const (
	token  = "your-api-token"
	userID = "userID"
)

func main() {
	api := slack.New(token)
	user, err := api.UsersInfo(userID)
	if err != nil {
		panic(err)
	}
	fmt.Println(user.Name, user.Profile.Email)
}
