package backend

import (
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	nsq "github.com/nsqio/go-nsq"
	"github.com/sensu/sensu-go/testing/util"
	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestHTTPListener(t *testing.T) {
	util.WithTempDir(func(path string) {
		ports := make([]int, 4)
		err := util.RandomPorts(ports)
		if err != nil {
			log.Panic(err)
		}
		clURL := fmt.Sprintf("http://127.0.0.1:%d", ports[0])
		apURL := fmt.Sprintf("http://127.0.0.1:%d", ports[1])
		agentPort := ports[2]
		apiPort := ports[3]
		initCluster := fmt.Sprintf("default=%s", apURL)
		fmt.Println(initCluster)

		b, err := NewBackend(&Config{
			AgentPort:           agentPort,
			APIPort:             apiPort,
			StateDir:            path,
			EtcdClientListenURL: clURL,
			EtcdPeerListenURL:   apURL,
			EtcdInitialCluster:  initCluster,
		})
		assert.NoError(t, err)
		if err != nil {
			assert.FailNow(t, "failed to start backend")
		}

		err = b.Run()
		assert.NoError(t, err)

		for i := 0; i < 5; i++ {
			conn, derr := net.Dial("tcp", fmt.Sprintf("localhost:%d", agentPort))
			if derr != nil {
				fmt.Println("Waiting for backend to start")
				time.Sleep(time.Duration(i) * time.Second)
			} else {
				conn.Close()
				continue
			}
		}

		client, err := transport.Connect(fmt.Sprintf("ws://localhost:%d/", agentPort))
		assert.NoError(t, err)
		assert.NotNil(t, client)

		msg := &transport.Message{
			Type:    types.AgentHandshakeType,
			Payload: []byte("{}"),
		}
		err = client.Send(msg)
		assert.NoError(t, err)
		resp, err := client.Receive()
		assert.NoError(t, err)
		assert.Equal(t, types.BackendHandshakeType, resp.Type)

		assert.NoError(t, client.Close())

		// Test NSQD
		nsqCfg := nsq.NewConfig()
		producer, err := nsq.NewProducer("127.0.0.1:4150", nsqCfg)
		assert.NoError(t, err)

		err = producer.Publish("test_topic", []byte("{}"))
		assert.NoError(t, err)
		b.Stop()
	})
}
