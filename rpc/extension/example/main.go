// This program is a sensu extension. It runs a gRPC server on its specified
// port. The extension filters all events that have a successful check and
// logs all events that have a failing check.
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"

	"github.com/sensu/sensu-go/rpc/extension"
	"github.com/sensu/sensu-go/types"
)

var (
	port = flag.Int("port", 31000, "port to run extension server on")
)

func LastCheckStatusZero(event *types.Event) (bool, error) {
	if event.Check == nil {
		// Filter metrics events as well.
		return true, nil
	}
	return event.Check.Status == 0, nil
}

func FailingCheck(event *types.Event, mutated []byte) error {
	log.Printf("entity %q: %s is failing", event.Entity.ID, event.Check.Name)
	return nil
}

func main() {
	flag.Parse()
	ext := extension.New().Filter(LastCheckStatusZero).Handle(FailingCheck)

	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", *port))
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	server := extension.NewServer(ext, ln)
	defer server.Stop()

	if err := server.Start(); err != nil {
		log.Fatal(err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	<-c
	log.Println("Shutting down...")
}
