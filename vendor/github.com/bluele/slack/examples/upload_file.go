package main

import (
	"fmt"
	"github.com/bluele/slack"
	"path/filepath"
)

// Please change these values to suit your environment
const (
	token          = "your-api-token"
	channelName    = "general"
	uploadFilePath = "./assets/test.txt"
)

func main() {
	api := slack.New(token)
	channel, err := api.FindChannelByName(channelName)
	if err != nil {
		panic(err)
	}

	err = api.FilesUpload(&slack.FilesUploadOpt{
		Filepath: uploadFilePath,
		Filetype: "text",
		Filename: filepath.Base(uploadFilePath),
		Title:    "upload test",
		Channels: []string{channel.Id},
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("Completed file upload.")
}
