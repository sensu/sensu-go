// mockextension is a sensu extension service whose behaviour can be controlled
// through environment variables.
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sensu/sensu-go/rpc"
	"google.golang.org/grpc"
)

// These are all for controlling output in the extension rpc methods.
const (
	SENSU_E2E_HANDLE_RESPONSE = "SENSU_E2E_HANDLE_RESPONSE"
	SENSU_E2E_HANDLE_ERROR    = "SENSU_E2E_HANDLE_ERROR"
	SENSU_E2E_MUTATE_RESPONSE = "SENSU_E2E_MUTATE_RESPONSE"
	SENSU_E2E_MUTATE_ERROR    = "SENSU_E2E_MUTATE_ERROR"
	SENSU_E2E_FILTER_RESPONSE = "SENSU_E2E_FILTER_RESPONSE"
	SENSU_E2E_FILTER_ERROR    = "SENSU_E2E_FILTER_ERROR"
)

var (
	port = flag.Int("port", 31337, "port the extension is running on")

	respondedOnce sync.Once
	respondedWg   sync.WaitGroup
	responded     = func() {
		respondedOnce.Do(func() {
			respondedWg.Done()
		})
	}

	lastError atomic.Value
)

func init() {
	respondedWg.Add(1)
	lastError.Store(false)
}

type extension struct{}

// processEvent processes an incoming event. It is called by the exported
// extension methods.
// processEvent uses printing to standard out as a method of one-way
// communication to the outside world. Users are expected to synchronize on
// its output.
func (e *extension) processEvent(ctx context.Context, req interface{}, errVar, respVar string, respVal interface{}) (err error) {
	defer func() {
		if err != nil {
			fmt.Print(err)
		}
	}()
	defer responded()

	if rawErr := os.Getenv(errVar); rawErr != "" {
		lastError.Store(true)
		return errors.New(rawErr)
	}

	rawResp := os.Getenv(respVar)
	if err := json.NewDecoder(strings.NewReader(rawResp)).Decode(respVal); err != nil {
		lastError.Store(true)
		return fmt.Errorf("error decoding response: %s", err)
	}
	fmt.Print(rawResp)
	return nil
}

func (e *extension) HandleEvent(ctx context.Context, req *rpc.HandleEventRequest) (resp *rpc.HandleEventResponse, err error) {
	var response rpc.HandleEventResponse
	return &response, e.processEvent(ctx, req, SENSU_E2E_HANDLE_ERROR, SENSU_E2E_HANDLE_RESPONSE, &response)
}

func (e *extension) MutateEvent(ctx context.Context, req *rpc.MutateEventRequest) (resp *rpc.MutateEventResponse, err error) {
	var response rpc.MutateEventResponse
	return &response, e.processEvent(ctx, req, SENSU_E2E_MUTATE_ERROR, SENSU_E2E_MUTATE_RESPONSE, &response)
}

func (e *extension) FilterEvent(ctx context.Context, req *rpc.FilterEventRequest) (resp *rpc.FilterEventResponse, err error) {
	var response rpc.FilterEventResponse
	return &response, e.processEvent(ctx, req, SENSU_E2E_FILTER_ERROR, SENSU_E2E_FILTER_RESPONSE, &response)
}

func printEnv() {
	p := fmt.Fprintln
	o := os.Stderr
	p(o, "SENSU_E2E_HANDLE_RESPONSE=", os.Getenv(SENSU_E2E_HANDLE_RESPONSE))
	p(o, "SENSU_E2E_HANDLE_ERROR=", os.Getenv(SENSU_E2E_HANDLE_ERROR))
	p(o, "SENSU_E2E_MUTATE_RESPONSE=", os.Getenv(SENSU_E2E_MUTATE_RESPONSE))
	p(o, "SENSU_E2E_MUTATE_ERROR=", os.Getenv(SENSU_E2E_MUTATE_ERROR))
	p(o, "SENSU_E2E_FILTER_RESPONSE=", os.Getenv(SENSU_E2E_FILTER_RESPONSE))
	p(o, "SENSU_E2E_FILTER_ERROR=", os.Getenv(SENSU_E2E_FILTER_ERROR))
}

func main() {
	flag.Parse()
	printEnv()
	// nb: using localhost:%d here does not work reliably.
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", *port))
	if err != nil {
		log.Fatal(err)
	}
	server := grpc.NewServer()
	rpc.RegisterExtensionServer(server, &extension{})
	go func() {
		if err := server.Serve(ln); err != nil {
			log.Fatal(err)
		}
	}()
	go func() {
		<-time.After(15 * time.Second)
		log.Fatal("service did not respond within 15 seconds")
	}()
	fmt.Println("ready")
	respondedWg.Wait()
	if err := lastError.Load().(bool); err {
		os.Exit(1)
	}
	os.Exit(0)
}
