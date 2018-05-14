package e2e

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/coreos/etcd/pkg/fileutil"
	"github.com/sensu/sensu-go/testing/testutil"
)

var (
	backend                              *backendProcess
	agentPortCounter                     int64 = 20000
	agentPath, backendPath, sensuctlPath string
	binDir                               = filepath.Join("..", "..", "bin")
	toolsDir                             = filepath.Join(binDir, "tools")
)

func TestMain(m *testing.M) {
	flag.Parse()

	agentBin := testutil.CommandPath("sensu-agent")
	backendBin := testutil.CommandPath("sensu-backend")
	sensuctlBin := testutil.CommandPath("sensuctl")

	agentPath = filepath.Join(binDir, agentBin)
	backendPath = filepath.Join(binDir, backendBin)
	sensuctlPath = filepath.Join(binDir, sensuctlBin)

	if !fileutil.Exist(agentPath) {
		fmt.Println("missing agent binary: ", agentPath)
		os.Exit(1)
	}

	if !fileutil.Exist(backendPath) {
		fmt.Println("missing backend binary: ", backendPath)
		os.Exit(1)
	}

	if !fileutil.Exist(sensuctlPath) {
		fmt.Println("missing backend binary: ", backendPath)
		os.Exit(1)
	}

	status := func() (status int) {
		var cleanup func()
		var err error
		backend, cleanup, err = newDefaultBackend()
		if err != nil {
			log.Println(err)
			return 1
		}

		defer func() {
			e := recover()
			cleanup()
			if e != nil {
				panic(e)
			}
		}()

		if err := backend.Start(); err != nil {
			log.Println(err)
			return 1
		}

		// Make sure the backend is ready
		isOnline := waitForBackend(backend.HTTPURL)
		if !isOnline {
			log.Println("the backend never became ready in a timely fashion")
			return 1
		}

		return m.Run()
	}()

	os.Exit(status)
}
