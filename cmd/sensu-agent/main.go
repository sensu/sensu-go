package main

import "github.com/sensu/sensu-go/agent/cmd"

func main() {
	if err := cmd.Execute(); err != nil {
		logger.WithError(err).Fatal("error executing sensu-agent")
	}
}
