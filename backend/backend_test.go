package backend

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"testing"
	"time"

	nsq "github.com/nsqio/go-nsq"
	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func testWithTempDir(t *testing.T, f func(string)) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "sensu")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer os.RemoveAll(tmpDir)
	f(tmpDir)
}

func TestHTTPListener(t *testing.T) {
	testWithTempDir(t, func(path string) {
		b, err := NewBackend(&Config{
			Port:     31337,
			StateDir: path,
		})
		assert.NoError(t, err)

		err = b.Run()
		assert.NoError(t, err)

		for i := 0; i < 5; i++ {
			conn, err := net.Dial("tcp", "localhost:31337")
			if err != nil {
				fmt.Println("Waiting for backend to start")
				time.Sleep(time.Duration(i) * time.Second)
			} else {
				conn.Close()
				continue
			}
		}

		client, err := transport.Connect("ws://localhost:31337")
		assert.NoError(t, err)
		assert.NotNil(t, client)

		err = client.Send(context.Background(), types.AgentHandshakeType, []byte("{}"))
		assert.NoError(t, err)
		mt, _, err := client.Receive(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, types.BackendHandshakeType, mt)

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
