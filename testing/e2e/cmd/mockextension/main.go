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
	"os/signal"
	"reflect"
	"strings"

	"github.com/sensu/sensu-go/rpc"
	"google.golang.org/grpc"
)

// These are all for setting expectations and controlling output in the
// extension rpc methods.
const (
	SENSU_E2E_HANDLE_REQUEST  = "SENSU_E2E_HANDLE_REQUEST"
	SENSU_E2E_HANDLE_RESPONSE = "SENSU_E2E_HANDLE_RESPONSE"
	SENSU_E2E_HANDLE_ERROR    = "SENSU_E2E_HANDLE_ERROR"
	SENSU_E2E_MUTATE_REQUEST  = "SENSU_E2E_MUTATE_REQUEST"
	SENSU_E2E_MUTATE_RESPONSE = "SENSU_E2E_MUTATE_RESPONSE"
	SENSU_E2E_MUTATE_ERROR    = "SENSU_E2E_MUTATE_ERROR"
	SENSU_E2E_FILTER_REQUEST  = "SENSU_E2E_FILTER_REQUEST"
	SENSU_E2E_FILTER_RESPONSE = "SENSU_E2E_FILTER_RESPONSE"
	SENSU_E2E_FILTER_ERROR    = "SENSU_E2E_FILTER_ERROR"
)

var (
	port = flag.Int("port", 31337, "port the extension is running on")
)

type extension struct{}

func (e *extension) HandleEvent(ctx context.Context, req *rpc.HandleEventRequest) (*rpc.HandleEventResponse, error) {
	if rawReq := os.Getenv(SENSU_E2E_HANDLE_REQUEST); rawReq != "" {
		var expReq rpc.HandleEventRequest
		if err := json.NewDecoder(strings.NewReader(rawReq)).Decode(&expReq); err != nil {
			return nil, fmt.Errorf("error decoding SENSU_E2E_HANDLE_REQUEST: %s", err)
		}
		if !reflect.DeepEqual(req, expReq) {
			return nil, fmt.Errorf("expected request did not match actual: got %+v, want %+v", req, expReq)
		}
	}
	if rawErr := os.Getenv(SENSU_E2E_HANDLE_ERROR); rawErr != "" {
		return nil, errors.New(rawErr)
	}
	var resp rpc.HandleEventResponse
	if rawResp := os.Getenv(SENSU_E2E_HANDLE_RESPONSE); rawResp != "" {
		if err := json.NewDecoder(strings.NewReader(rawResp)).Decode(&rawResp); err != nil {
			return nil, fmt.Errorf("error decoding SENSU_E2E_HANDLE_RESPONSE: %s", err)
		}
	}
	return &resp, nil
}

func (e *extension) MutateEvent(ctx context.Context, req *rpc.MutateEventRequest) (*rpc.MutateEventResponse, error) {
	if rawReq := os.Getenv(SENSU_E2E_MUTATE_REQUEST); rawReq != "" {
		var expReq rpc.MutateEventRequest
		if err := json.NewDecoder(strings.NewReader(rawReq)).Decode(&expReq); err != nil {
			return nil, fmt.Errorf("error decoding SENSU_E2E_MUTATE_REQUEST: %s", err)
		}
		if !reflect.DeepEqual(req, expReq) {
			return nil, fmt.Errorf("expected request did not match actual: got %+v, want %+v", req, expReq)
		}
	}
	if rawErr := os.Getenv(SENSU_E2E_MUTATE_ERROR); rawErr != "" {
		return nil, errors.New(rawErr)
	}
	var resp rpc.MutateEventResponse
	if rawResp := os.Getenv(SENSU_E2E_MUTATE_RESPONSE); rawResp != "" {
		if err := json.NewDecoder(strings.NewReader(rawResp)).Decode(&rawResp); err != nil {
			return nil, fmt.Errorf("error decoding SENSU_E2E_MUTATE_RESPONSE: %s", err)
		}
	}
	return &resp, nil
}

func (e *extension) FilterEvent(ctx context.Context, req *rpc.FilterEventRequest) (*rpc.FilterEventResponse, error) {
	if rawReq := os.Getenv(SENSU_E2E_FILTER_REQUEST); rawReq != "" {
		var expReq rpc.FilterEventRequest
		if err := json.NewDecoder(strings.NewReader(rawReq)).Decode(&expReq); err != nil {
			return nil, fmt.Errorf("error decoding SENSU_E2E_FILTER_REQUEST: %s", err)
		}
		if !reflect.DeepEqual(req, expReq) {
			return nil, fmt.Errorf("expected request did not match actual: got %+v, want %+v", req, expReq)
		}
	}
	if rawErr := os.Getenv(SENSU_E2E_FILTER_ERROR); rawErr != "" {
		return nil, errors.New(rawErr)
	}
	var resp rpc.FilterEventResponse
	if rawResp := os.Getenv(SENSU_E2E_FILTER_RESPONSE); rawResp != "" {
		if err := json.NewDecoder(strings.NewReader(rawResp)).Decode(&rawResp); err != nil {
			return nil, fmt.Errorf("error decoding SENSU_E2E_FILTER_RESPONSE: %s", err)
		}
	}
	return &resp, nil
}

func printEnv() {
	p := fmt.Println
	p("SENSU_E2E_HANDLE_REQUEST=", os.Getenv("SENSU_E2E_HANDLE_REQUEST"))
	p("SENSU_E2E_HANDLE_RESPONSE=", os.Getenv("SENSU_E2E_HANDLE_RESPONSE"))
	p("SENSU_E2E_HANDLE_ERROR=", os.Getenv("SENSU_E2E_HANDLE_ERROR"))
	p("SENSU_E2E_MUTATE_REQUEST=", os.Getenv("SENSU_E2E_MUTATE_REQUEST"))
	p("SENSU_E2E_MUTATE_RESPONSE=", os.Getenv("SENSU_E2E_MUTATE_RESPONSE"))
	p("SENSU_E2E_MUTATE_ERROR=", os.Getenv("SENSU_E2E_MUTATE_ERROR"))
	p("SENSU_E2E_FILTER_REQUEST=", os.Getenv("SENSU_E2E_FILTER_REQUEST"))
	p("SENSU_E2E_FILTER_RESPONSE=", os.Getenv("SENSU_E2E_FILTER_RESPONSE"))
	p("SENSU_E2E_FILTER_EVENT_ERROR=", os.Getenv("SENSU_E2E_FILTER_EVENT_ERROR"))
}

func main() {
	flag.Parse()
	printEnv()
	ln, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", *port))
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
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	os.Exit(0)
}
