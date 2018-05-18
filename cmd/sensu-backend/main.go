package main

import (
	_ "net/http/pprof"
	"os"

	"github.com/sensu/sensu-go/backend"
	"github.com/sensu/sensu-go/backend/cmd"
)

func main() {
	backend := &backend.Backend{}
	if err := cmd.Execute(backend); err != nil {
		os.Exit(1)
	}
}
