package e2e

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"testing"

	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func init() {
	err := exec.Command("go", "install", "github.com/sensu/sensu-go/testing/e2e/cmd/mockextension").Run()
	if err != nil {
		log.Fatal(err)
	}
}

func newExtension(t *testing.T, port int, env map[string]string) func() (string, error) {
	cmd := exec.Command("mockextension", "-port", fmt.Sprintf("%d", port))

	// Set up the environment
	for k, v := range env {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Divert stdout to a pipe
	out, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}

	// Start the mock extension
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}

	// Wait until the listener is ready
	ready := make([]byte, len("ready\n"))
	if _, err := out.Read(ready); err != nil {
		t.Fatal(err)
	}
	if r := string(ready); r != "ready\n" {
		t.Fatal(errors.New(r))
	}

	// The cleanup func reads all the remaining standard out
	// and waits for the command to terminate
	return func() (string, error) {
		buf, err := ioutil.ReadAll(out)
		if err != nil {
			return "", err
		}
		err = cmd.Wait()
		if exitErr, ok := err.(*exec.ExitError); ok {
			if len(exitErr.Stderr) != 0 {
				return string(exitErr.Stderr), err
			}
		}
		return string(buf), err
	}
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
		Name   string
		Type   string
		Env    map[string]string
		Event  *types.Event
		ExpErr bool
	}{
		{
			Name:  "no error",
			Type:  "handle",
			Event: handleEvt,
			Env: map[string]string{
				"SENSU_E2E_HANDLE_RESPONSE": "{}",
			},
		},
		{
			Name:  "RPC error",
			Type:  "handle",
			Event: handleEvt,
			Env: map[string]string{
				"SENSU_E2E_HANDLE_ERROR":    "i/o timeout",
				"SENSU_E2E_HANDLE_RESPONSE": "{}",
			},
			ExpErr: true,
		},
		{
			Name:  "logic error",
			Type:  "handle",
			Event: handleEvt,
			Env: map[string]string{
				"SENSU_E2E_HANDLE_RESPONSE": `{"error": "some error"}`,
			},
			ExpErr: true,
		},
		{
			Name:  "no error and not filtered",
			Type:  "filter",
			Event: filterEvt,
			Env: map[string]string{
				"SENSU_E2E_FILTER_RESPONSE": "{}",
			},
		},
		{
			Name:  "no error and filtered",
			Type:  "filter",
			Event: filterEvt,
			Env: map[string]string{
				"SENSU_E2E_FILTER_RESPONSE": `{"filtered": true}`,
			},
		},
		{
			Name:  "RPC error",
			Type:  "filter",
			Event: filterEvt,
			Env: map[string]string{
				"SENSU_E2E_FILTER_ERROR":    "i/o timeout",
				"SENSU_E2E_FILTER_RESPONSE": "{}",
			},
			ExpErr: true,
		},
		{
			Name:  "logic error",
			Type:  "filter",
			Event: filterEvt,
			Env: map[string]string{
				"SENSU_E2E_FILTER_RESPONSE": `{"error":"internal error"}`,
			},
			ExpErr: true,
		},
		{
			Name:  "no error",
			Type:  "mutate",
			Event: mutateEvt,
			Env: map[string]string{
				"SENSU_E2E_MUTATE_RESPONSE": `{"mutatedEvent":"{}"}`,
			},
		},
		{
			Name:  "RPC error",
			Type:  "mutate",
			Event: mutateEvt,
			Env: map[string]string{
				"SENSU_E2E_MUTATE_ERROR":    "i/o timeout",
				"SENSU_E2E_MUTATE_RESPONSE": "{}",
			},
			ExpErr: true,
		},
		{
			Name:  "logic error",
			Type:  "mutate",
			Event: mutateEvt,
			Env: map[string]string{
				"SENSU_E2E_MUTATE_RESPONSE": `{"error":"internal error"}`,
			},
			ExpErr: true,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s %s", test.Type, test.Name), func(t *testing.T) {
			cleanup := newExtension(t, ports[0], test.Env)

			body, _ := json.Marshal(test.Event)
			resp, err := client.R().SetBody(body).Put(
				fmt.Sprintf("/events/%s/%s", test.Event.Entity.ID, test.Event.Check.Name))
			if err != nil {
				t.Fatal(err)
			}
			if status := resp.StatusCode(); status >= 400 {
				t.Fatalf("bad status: %d", status)
			}

			out, err := cleanup()
			if err != nil && !test.ExpErr {
				fmt.Println(out)
				t.Fatal(err)
			}

			switch test.Type {
			case "handle":
				if err == nil {
					assert.Equal(t, test.Env["SENSU_E2E_HANDLE_RESPONSE"], out)
				} else {
					assert.Equal(t, test.Env["SENSU_E2E_HANDLE_ERROR"], out)
				}
			case "filter":
				if err == nil {
					assert.Equal(t, test.Env["SENSU_E2E_FILTER_RESPONSE"], out)
				} else {
					assert.Equal(t, test.Env["SENSU_E2E_FILTER_ERROR"], out)
				}
			case "mutate":
				if err == nil {
					assert.Equal(t, test.Env["SENSU_E2E_MUTATE_RESPONSE"], out)
				} else {
					assert.Equal(t, test.Env["SENSU_E2E_MUTATE_ERROR"], out)
				}
			default:
				t.Fatalf("invalid test type: %q", test.Type)
			}
		})
	}
}
