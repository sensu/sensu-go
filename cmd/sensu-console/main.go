package main

import (
	"log"

	"github.com/sensu/sensu-go/sdk/cmd"
)

func main() {
	command := cmd.SDKCommand()
	if err := command.Execute(); err != nil {
		log.Fatal(err)
	}
}
