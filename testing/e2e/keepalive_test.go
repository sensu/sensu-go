package e2e

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sensu/sensu-go/cli/client"
	"github.com/sensu/sensu-go/cli/client/config/basic"
	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

// TODO(greg): Yeah, this is really just one enormous test for all e2e stuff.
// I'd love to see this organized better.
func TestAgentKeepalives(t *testing.T) {
	ports := make([]int, 5)
	err := testutil.RandomPorts(ports)
	if err != nil {
		log.Fatal(err)
	}

	tmpDir, err := ioutil.TempDir(os.TempDir(), "sensu")
	if err != nil {
		log.Panic(err)
	}
	defer os.RemoveAll(tmpDir)

	etcdClientURL := fmt.Sprintf("http://127.0.0.1:%d", ports[0])
	etcdPeerURL := fmt.Sprintf("http://127.0.0.1:%d", ports[1])
	apiPort := ports[2]
	agentPort := ports[3]
	dashboardPort := ports[4]
	backendWSURL := fmt.Sprintf("ws://127.0.0.1:%d/", agentPort)
	backendHTTPURL := fmt.Sprintf("http://127.0.0.1:%d", apiPort)
	initialCluster := fmt.Sprintf("default=%s", etcdPeerURL)

	bep := &backendProcess{
		AgentHost:               "127.0.0.1",
		AgentPort:               agentPort,
		APIHost:                 "127.0.0.1",
		APIPort:                 apiPort,
		DashboardHost:           "127.0.0.1",
		DashboardPort:           dashboardPort,
		StateDir:                tmpDir,
		EtcdClientURL:           etcdClientURL,
		EtcdPeerURL:             etcdPeerURL,
		EtcdInitialCluster:      initialCluster,
		EtcdInitialClusterState: "new",
		EtcdName:                "default",
	}

	err = bep.Start()
	if err != nil {
		log.Panic(err)
	}

	ap := &agentProcess{
		// testing the StringSlice for backend-url and the backend selector.
		BackendURLs: []string{backendWSURL, backendWSURL},
		AgentID:     "TestKeepalives",
	}

	backendHealthy := false
	for i := 0; i < 10; i++ {
		resp, getErr := http.Get(fmt.Sprintf("%s/health", backendHTTPURL))
		if getErr != nil {
			log.Println("backend not ready, sleeping...")
			time.Sleep(1 * time.Second)
			continue
		}
		resp.Body.Close()
		if resp.StatusCode != 200 && resp.StatusCode != 401 {
			log.Printf("backend returned non-200/401 status code: %d\n", resp.StatusCode)
			time.Sleep(1 * time.Second)
			continue
		}
		backendHealthy = true
	}

	assert.True(t, backendHealthy)

	// Create an authenticated HTTP Sensu client
	clientConfig := &basic.Config{
		Cluster: basic.Cluster{
			APIUrl: backendHTTPURL,
		},
	}
	sensuClient := client.New(clientConfig)
	tokens, _ := sensuClient.CreateAccessToken(backendHTTPURL, "admin", "P@ssw0rd!")
	clientConfig.Cluster.Tokens = tokens

	err = ap.Start()
	assert.NoError(t, err)

	defer func() {
		bep.Kill()
		ap.Kill()
	}()

	// Give it a second to make sure we've sent a keepalive.
	time.Sleep(5 * time.Second)

	// Retrieve the entitites
	entities, err := sensuClient.ListEntities("*")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(entities))
	assert.Equal(t, "TestKeepalives", entities[0].ID)
	assert.Equal(t, "agent", entities[0].Class)
	assert.NotEmpty(t, entities[0].System.Hostname)
	assert.NotZero(t, entities[0].LastSeen)

	// Create a check
	check := &types.CheckConfig{
		Name:          "testcheck",
		Command:       "echo output",
		Interval:      1,
		Subscriptions: []string{"test"},
		Environment:   "default",
		Organization:  "default",
	}
	err = sensuClient.CreateCheck(check)
	assert.NoError(t, err)

	// Retrieve the check
	_, err = sensuClient.FetchCheck(check.Name)
	assert.NoError(t, err)

	falsePath := testutil.CommandPath(filepath.Join(binDir, "false"))
	falseAbsPath, err := filepath.Abs(falsePath)
	assert.NoError(t, err)
	assert.NotEmpty(t, falseAbsPath)

	check = &types.CheckConfig{
		Name:          "testcheck2",
		Command:       falseAbsPath,
		Interval:      1,
		Subscriptions: []string{"test"},
		Environment:   "default",
		Organization:  "default",
	}
	err = sensuClient.CreateCheck(check)
	assert.NoError(t, err)

	time.Sleep(30 * time.Second)

	// At this point, we should have 21 failing status codes for testcheck2
	event, err := sensuClient.FetchEvent(ap.AgentID, check.Name)
	assert.NoError(t, err)
	assert.NotNil(t, event)
	assert.NotNil(t, event.Check)
	assert.NotNil(t, event.Entity)
	assert.Equal(t, "TestKeepalives", event.Entity.ID)
	assert.Equal(t, "testcheck2", event.Check.Config.Name)
	// TODO(greg): ensure results are as expected.
}
