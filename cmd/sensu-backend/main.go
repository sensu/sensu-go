package main

import (
	_ "net/http/pprof"

	"github.com/sensu/sensu-go/backend"
	"github.com/sensu/sensu-go/backend/cmd"
)

func main() {
	initializeFn := backend.Initialize
	if err := cmd.Execute(initializeFn); err != nil {
		logger.Fatal(err)
	}
}
