package main

import (
	_ "net/http/pprof"

	"github.com/sensu/sensu-go/backend"
	"github.com/sensu/sensu-go/backend/cmd"
)

func main() {
	if err := cmd.Execute(backend.Initialize); err != nil {
		logger.Fatal(err)
	}
}
