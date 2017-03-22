package e2e

import (
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/coreos/etcd/pkg/fileutil"
)

var binDir string

func TestMain(m *testing.M) {
	flag.StringVar(&binDir, "bin-dir", "../../bin", "directory containing sensu binaries")
	flag.Parse()

	if !fileutil.Exist(binDir+"/sensu-agent") || !fileutil.Exist(binDir+"/sensu-backend") {
		fmt.Println("missing binaries")
		os.Exit(1)
	}

	os.Exit(m.Run())
}
