package main

import (
	_ "net/http/pprof"

	"github.com/sensu/sensu-go/backend"
	"github.com/sensu/sensu-go/backend/cmd"
)

func main() {
	backend := &backend.Backend{}
	if err := cmd.Execute(backend); err != nil {
		logger.Fatal(err)
	}
}
