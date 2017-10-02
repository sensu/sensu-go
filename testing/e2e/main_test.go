package e2e

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/coreos/etcd/pkg/fileutil"
	"github.com/sensu/sensu-go/testing/testutil"
)

var binDir, agentPath, backendPath, sensuctlPath string

func TestMain(m *testing.M) {
	flag.StringVar(&binDir, "bin-dir", "../../bin", "directory containing sensu binaries")
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

	os.Exit(m.Run())
}
