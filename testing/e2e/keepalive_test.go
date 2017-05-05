package e2e

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sensu/sensu-go/testing/util"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

// TODO(greg): Yeah, this is really just one enormous test for all e2e stuff.
// I'd love to see this organized better.
func TestAgentKeepalives(t *testing.T) {
	ports := make([]int, 4)
	err := util.RandomPorts(ports)
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
	backendWSURL := fmt.Sprintf("ws://127.0.0.1:%d/", agentPort)
	backendHTTPURL := fmt.Sprintf("http://127.0.0.1:%d", apiPort)
	initialCluster := fmt.Sprintf("default=%s", etcdPeerURL)

	bep := &backendProcess{
		APIPort:            apiPort,
		AgentPort:          agentPort,
		StateDir:           tmpDir,
		EtcdClientURL:      etcdClientURL,
		EtcdPeerURL:        etcdPeerURL,
		EtcdInitialCluster: initialCluster,
	}

	err = bep.Start()
	if err != nil {
		log.Panic(err)
	}

	ap := &agentProcess{
		BackendURL: backendWSURL,
		AgentID:    "TestKeepalives",
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
		if resp.StatusCode != 200 {
			log.Printf("backend returned non-200 status code: %d\n", resp.StatusCode)
			time.Sleep(1 * time.Second)
			continue
		}
		backendHealthy = true
	}

	assert.True(t, backendHealthy)

	err = ap.Start()
	assert.NoError(t, err)

	defer func() {
		bep.Kill()
		ap.Kill()
	}()

	// Give it a second to make sure we've sent a keepalive.
	time.Sleep(5 * time.Second)

	resp, err := http.Get(fmt.Sprintf("%s/entities", backendHTTPURL))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()
	entities := []*types.Entity{}
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	err = json.Unmarshal(bodyBytes, &entities)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(entities))
	assert.Equal(t, "TestKeepalives", entities[0].ID)
	assert.Equal(t, "agent", entities[0].Class)
	assert.NotEmpty(t, entities[0].System.Hostname)

	check := &types.Check{
		Name:          "testcheck",
		Command:       "echo output",
		Interval:      1,
		Subscriptions: []string{"test"},
	}
	checkBytes, err := json.Marshal(check)
	assert.NoError(t, err)
	resp, err = http.Post(fmt.Sprintf("%s/checks/testcheck", backendHTTPURL), "application/json", bytes.NewBuffer(checkBytes))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NoError(t, err)
	resp, err = http.Get(fmt.Sprintf("%s/checks/testcheck", backendHTTPURL))
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	falsePath := util.CommandPath(filepath.Join(binDir, "false"))
	falseAbsPath, err := filepath.Abs(falsePath)
	assert.NoError(t, err)
	assert.NotEmpty(t, falseAbsPath)

	check = &types.Check{
		Name:          "testcheck2",
		Command:       falseAbsPath,
		Interval:      1,
		Subscriptions: []string{"test"},
	}
	checkBytes, err = json.Marshal(check)
	assert.NoError(t, err)
	resp, err = http.Post(fmt.Sprintf("%s/checks/testcheck2", backendHTTPURL), "application/json", bytes.NewBuffer(checkBytes))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NoError(t, err)

	time.Sleep(30 * time.Second)

	// At this point, we should have 21 failing status codes for testcheck2
	resp, err = http.Get(fmt.Sprintf("%s/events/TestKeepalives/testcheck2", backendHTTPURL))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.NoError(t, err)

	eventBytes, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	resp.Body.Close()
	event := &types.Event{}
	json.Unmarshal(eventBytes, event)
	assert.NotNil(t, event)
	assert.NotNil(t, event.Check)
	assert.NotNil(t, event.Entity)
	assert.Equal(t, "TestKeepalives", event.Entity.ID)
	assert.Equal(t, "testcheck2", event.Check.Name)
	// TODO(greg): ensure results are as expected.
}
