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

func (e *extension) HandleEvent(ctx context.Context, req *rpc.HandleEventRequest) (resp *rpc.HandleEventResponse, err error) {
	defer func() {
		if err != nil {
			fmt.Print(err)
		}
	}()
	defer responded()
	if rawErr := os.Getenv(SENSU_E2E_HANDLE_ERROR); rawErr != "" {
		lastError.Store(true)
		return nil, errors.New(rawErr)
	}
	var response rpc.HandleEventResponse
	rawResp := os.Getenv(SENSU_E2E_HANDLE_RESPONSE)
	if err := json.NewDecoder(strings.NewReader(rawResp)).Decode(&response); err != nil {
		lastError.Store(true)
		return nil, fmt.Errorf("error decoding SENSU_E2E_HANDLE_RESPONSE: %s", err)
	}
	fmt.Print(rawResp)
	return &response, nil
}

func (e *extension) MutateEvent(ctx context.Context, req *rpc.MutateEventRequest) (resp *rpc.MutateEventResponse, err error) {
	defer func() {
		if err != nil {
			fmt.Print(err)
		}
	}()
	defer responded()
	if rawErr := os.Getenv(SENSU_E2E_MUTATE_ERROR); rawErr != "" {
		lastError.Store(true)
		return nil, errors.New(rawErr)
	}
	var response rpc.MutateEventResponse
	rawResp := os.Getenv(SENSU_E2E_MUTATE_RESPONSE)
	if err := json.NewDecoder(strings.NewReader(rawResp)).Decode(&response); err != nil {
		lastError.Store(true)
		return nil, fmt.Errorf("error decoding SENSU_E2E_MUTATE_RESPONSE: %s", err)
	}
	fmt.Print(rawResp)
	return &response, nil
}

func (e *extension) FilterEvent(ctx context.Context, req *rpc.FilterEventRequest) (resp *rpc.FilterEventResponse, err error) {
	defer func() {
		if err != nil {
			fmt.Print(err)
		}
	}()
	defer responded()
	if rawErr := os.Getenv(SENSU_E2E_FILTER_ERROR); rawErr != "" {
		lastError.Store(true)
		fmt.Fprintln(os.Stderr, rawErr)
		return nil, errors.New(rawErr)
	}
	var response rpc.FilterEventResponse
	rawResp := os.Getenv(SENSU_E2E_FILTER_RESPONSE)
	if err := json.NewDecoder(strings.NewReader(rawResp)).Decode(&response); err != nil {
		lastError.Store(true)
		return nil, fmt.Errorf("error decoding SENSU_E2E_FILTER_RESPONSE: %s", err)
	}
	fmt.Print(rawResp)
	return &response, nil
}

func printEnv() {
	p := fmt.Fprintln
	o := os.Stderr
	p(o, "SENSU_E2E_HANDLE_RESPONSE=", os.Getenv("SENSU_E2E_HANDLE_RESPONSE"))
	p(o, "SENSU_E2E_HANDLE_ERROR=", os.Getenv("SENSU_E2E_HANDLE_ERROR"))
	p(o, "SENSU_E2E_MUTATE_RESPONSE=", os.Getenv("SENSU_E2E_MUTATE_RESPONSE"))
	p(o, "SENSU_E2E_MUTATE_ERROR=", os.Getenv("SENSU_E2E_MUTATE_ERROR"))
	p(o, "SENSU_E2E_FILTER_RESPONSE=", os.Getenv("SENSU_E2E_FILTER_RESPONSE"))
	p(o, "SENSU_E2E_FILTER_EVENT_ERROR=", os.Getenv("SENSU_E2E_FILTER_EVENT_ERROR"))
}

func main() {
	flag.Parse()
	printEnv()
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
		log.Fatal("service did not respond within 10 seconds")
	}()
	fmt.Println("ready")
	respondedWg.Wait()
	if err := lastError.Load().(bool); err {
		os.Exit(1)
	}
	os.Exit(0)
}
