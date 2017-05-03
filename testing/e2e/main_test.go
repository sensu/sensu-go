package e2e

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/coreos/etcd/pkg/fileutil"
	"github.com/sensu/sensu-go/testing/util"
)

var binDir string

func TestMain(m *testing.M) {
	flag.StringVar(&binDir, "bin-dir", "../../bin", "directory containing sensu binaries")
	flag.Parse()

	agentBin := util.CommandPath("sensu-agent")
	backendBin := util.CommandPath("sensu-backend")

	agentPath := filepath.Join(binDir, agentBin)
	backendPath := filepath.Join(binDir, backendBin)

	if !fileutil.Exist(agentPath) {
		fmt.Println("missing agent binary: ", agentPath)
		os.Exit(1)
	}

	if !fileutil.Exist(backendPath) {
		fmt.Println("missing backend binary: ", backendPath)
		os.Exit(1)
	}

	os.Exit(m.Run())
}
