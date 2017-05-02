package e2e

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/coreos/etcd/pkg/fileutil"
)

var binDir string

func TestMain(m *testing.M) {
	flag.StringVar(&binDir, "bin-dir", "../../bin", "directory containing sensu binaries")
	flag.Parse()

	var (
		agentBin string
		backendBin string
	)

	switch runtime.GOOS {
	case "windows":
		agentBin = "sensu-agent.exe"
		backendBin = "sensu-backend.exe"
	default:
		agentBin = "sensu-agent"
		backendBin = "sensu-backend"
	}

	agentPath := filepath.Join(binDir, agentBin)
	backendPath := filepath.Join(binDir, backendBin)

	if !fileutil.Exist(agentPath) || !fileutil.Exist(backendPath) {
		fmt.Println("missing binaries")
		os.Exit(1)
	}

	os.Exit(m.Run())
}
