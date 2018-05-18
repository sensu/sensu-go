package main

import (
	_ "net/http/pprof"

	"github.com/sensu/sensu-go/backend/cmd"
	"github.com/sensu/sensu-go/backend/core"
)

func main() {
	backend := &core.Backend{}
	cmd.Execute(backend)
}
