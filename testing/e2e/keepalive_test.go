package e2e

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/sensu/sensu-go/testing/util"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

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

	// We do our debug/logging output here so that we don't panic down the line and
	// never see it. This is all pretty useful stuff. This also lets us shutdown our
	// child processes cleanly.
	defer func() {
		// We get vetshadow errors if we use err here, which is really damn
		// annoying.
		var dErr error
		bep.Kill()
		ap.Kill()

		b, dErr := ioutil.ReadAll(bep.Stderr)
		if dErr != nil {
			log.Panic(dErr)
		}
		fmt.Print(string(b))

		b, dErr = ioutil.ReadAll(ap.Stderr)
		if dErr != nil {
			log.Panic(dErr)
		}
		fmt.Print(string(b))
	}()

	// Give it a second to make sure we've sent a keepalive.
	time.Sleep(1 * time.Second)

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
	assert.NotEmpty(t, entities[0].System.OS)
	assert.NotEmpty(t, entities[0].System.Platform)
	assert.NotEmpty(t, entities[0].System.PlatformFamily)
	assert.NotEmpty(t, entities[0].System.PlatformVersion)
}
