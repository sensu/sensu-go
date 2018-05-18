package main

import (
	_ "net/http/pprof"

	"github.com/sensu/sensu-go/backend"
	"github.com/sensu/sensu-go/backend/cmd"
)

func main() {
	backend := &backend.Backend{}
	cmd.Execute(backend)
}
