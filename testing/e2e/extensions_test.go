package e2e

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"strings"
	"sync"
	"testing"

	"github.com/sensu/sensu-go/rpc"
	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type mockExtension struct {
	setResponse string
	setError    string

	response string
	err      string

	reqType string
	server  *grpc.Server
	ln      net.Listener
	sync.WaitGroup
}

func newMockExtension(ln net.Listener, resp, err string) *mockExtension {
	ext := &mockExtension{
		setResponse: resp,
		setError:    err,
		ln:          ln,
	}
	ext.Add(1)
	ext.server = grpc.NewServer()
	rpc.RegisterExtensionServer(ext.server, ext)
	go ext.server.Serve(ln)
	return ext
}

func (e *mockExtension) Stop() {
	e.server.Stop()
	e.ln.Close()
}

// processEvent processes an incoming event. It is called by the exported
// extension methods.
// processEvent uses printing to standard out as a method of one-way
// communication to the outside world. Users are expected to synchronize on
// its output.
func (e *mockExtension) processEvent(ctx context.Context, req interface{}, respVal interface{}) (err error) {
	defer func() {
		if err != nil {
			e.err = err.Error()
		}
	}()
	defer e.Done()

	if e.setError != "" {
		return errors.New(e.setError)
	}

	if err := json.NewDecoder(strings.NewReader(e.setResponse)).Decode(respVal); err != nil {
		return fmt.Errorf("error decoding response: %s", err)
	}
	return nil
}

func (e *mockExtension) HandleEvent(ctx context.Context, req *rpc.HandleEventRequest) (resp *rpc.HandleEventResponse, err error) {
	var response rpc.HandleEventResponse
	e.reqType = "handle"
	return &response, e.processEvent(ctx, req, &response)
}

func (e *mockExtension) MutateEvent(ctx context.Context, req *rpc.MutateEventRequest) (resp *rpc.MutateEventResponse, err error) {
	var response rpc.MutateEventResponse
	e.reqType = "mutate"
	return &response, e.processEvent(ctx, req, &response)
}

func (e *mockExtension) FilterEvent(ctx context.Context, req *rpc.FilterEventRequest) (resp *rpc.FilterEventResponse, err error) {
	var response rpc.FilterEventResponse
	e.reqType = "filter"
	return &response, e.processEvent(ctx, req, &response)
}

func TestExtensions(t *testing.T) {
	t.Parallel()

	// Start the backend
	backend, cleanup := newBackend(t)
	defer cleanup()

	// Initializes client & sensuctl
	client := newSensuClient(backend.HTTPURL)
	sensuctl, cleanup := newSensuCtl(backend.HTTPURL, "default", "default", "admin", "P@ssw0rd!")
	defer cleanup()

	// Start the agent
	agentConfig := agentConfig{
		ID:          "TestExtensions",
		BackendURLs: []string{backend.WSURL},
	}
	_, cleanup = newAgent(agentConfig, sensuctl, t)
	defer cleanup()

	// Register the extension service
	ports := make([]int, 1)
	err := testutil.RandomPorts(ports)
	if err != nil {
		t.Fatal(err)
	}
	out, err := sensuctl.run("extension", "register", "extension1", fmt.Sprintf("127.0.0.1:%d", ports[0]))
	if err != nil {
		fmt.Println(string(out))
		t.Fatal(err)
	}

	// create a filter handler
	out, err = sensuctl.run("handler", "create", "filter1", "--filters", "extension1")
	if err != nil {
		fmt.Println(string(out))
		t.Fatal(err)
	}

	// create a mutator handler
	out, err = sensuctl.run("handler", "create", "mutator1", "-m", "extension1")
	if err != nil {
		fmt.Println(string(out))
		t.Fatal(err)
	}

	// Create a check
	check := types.FixtureCheckConfig("check1")
	check.Publish = false
	check.Interval = 1
	if err := client.CreateCheck(check); err != nil {
		t.Fatal(err)
	}

	// This event is meant to test HandleEvent
	handleEvt := types.FixtureEvent("TestExtensions", "check1")
	handleEvt.Check.Handlers = append(handleEvt.Check.Handlers, "extension1")

	// This event is meant to test FilterEvent
	filterEvt := types.FixtureEvent("TestExtensions", "check1")
	filterEvt.Check.Handlers = append(filterEvt.Check.Handlers, "filter1")

	// This event is meant to test MutateEvent
	mutateEvt := types.FixtureEvent("TestExtensions", "check1")
	mutateEvt.Check.Handlers = append(mutateEvt.Check.Handlers, "mutator1")

	tests := []struct {
		Name     string
		Type     string
		Response string
		Error    string
		Event    *types.Event
		ExpErr   bool
	}{
		{
			Name:     "no error",
			Type:     "handle",
			Event:    handleEvt,
			Response: "{}",
		},
		{
			Name:   "RPC error",
			Type:   "handle",
			Event:  handleEvt,
			Error:  "i/o timeout",
			ExpErr: true,
		},
		{
			Name:     "logic error",
			Type:     "handle",
			Event:    handleEvt,
			Response: `{"error": "some error"}`,
			ExpErr:   true,
		},
		{
			Name:     "no error and not filtered",
			Type:     "filter",
			Event:    filterEvt,
			Response: "{}",
		},
		{
			Name:     "no error and filtered",
			Type:     "filter",
			Event:    filterEvt,
			Response: `{"filtered": true}`,
		},
		{
			Name:   "RPC error",
			Type:   "filter",
			Event:  filterEvt,
			Error:  "i/o timeout",
			ExpErr: true,
		},
		{
			Name:     "logic error",
			Type:     "filter",
			Event:    filterEvt,
			Response: `{"error":"internal error"}`,
			ExpErr:   true,
		},
		{
			Name:     "no error",
			Type:     "mutate",
			Event:    mutateEvt,
			Response: `{"mutatedEvent":"{}"}`,
		},
		{
			Name:   "RPC error",
			Type:   "mutate",
			Event:  mutateEvt,
			Error:  "i/o timeout",
			ExpErr: true,
		},
		{
			Name:     "logic error",
			Type:     "mutate",
			Event:    mutateEvt,
			Response: `{"error":"internal error"}`,
			ExpErr:   true,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s %s", test.Type, test.Name), func(t *testing.T) {
			ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", ports[0]))
			if err != nil {
				t.Fatal(err)
			}
			ext := newMockExtension(ln, test.Response, test.Error)
			defer ext.Stop()

			body, _ := json.Marshal(test.Event)
			resp, err := client.R().SetBody(body).Put(
				fmt.Sprintf("/events/%s/%s", test.Event.Entity.ID, test.Event.Check.Name))
			if err != nil {
				t.Fatal(err)
			}
			if status := resp.StatusCode(); status >= 400 {
				t.Fatalf("bad status: %d", status)
			}

			ext.Wait()
			if ext.err != "" && !test.ExpErr {
				t.Fatal(ext.err)
			}

			assert.Equal(t, test.Error, ext.err)
			assert.Equal(t, test.Type, ext.reqType)
		})
	}
}
