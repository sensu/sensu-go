package testutil

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

var socketChan = make(chan net.Listener, 100)

func init() {
	// dirty hack to prevent port collisions
	go func() {
		for {
			ln, err := net.Listen("tcp4", "127.0.0.1:0")
			if err != nil {
				log.Println(err)
			}
			socketChan <- ln
		}
	}()
}

// TempDir provides a test with a temporary directory (under os.TempDir())
// returning the absolute path to the directory and a remove() function
// that should be deferred immediately after calling TempDir(t) to recursively
// delete the contents of the directory.
func TempDir(t *testing.T) (tmpDir string, remove func()) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "sensu")
	if err != nil {
		t.FailNow()
	}

	return tmpDir, func() { _ = os.RemoveAll(tmpDir) }
}

// RandomPorts reserves len(p) ports and assigns them to elements of p.
// It calls ReservePort repeatedly.
func RandomPorts(p []int) (err error) {
	for i := range p {
		ln := <-socketChan
		defer ln.Close()
		port, err := strconv.Atoi(strings.Split(ln.Addr().String(), ":")[1])
		if err != nil {
			return err
		}
		p[i] = port
	}
	return nil
}

// CleanOutput takes a string and strips extra characters that are not
// consistent across platforms.
func CleanOutput(s string) string {
	return strings.Replace(s, "\r", "", -1)
}

// CommandPath takes a path to a command and returns a corrected path
// for the operating system this is running on.
func CommandPath(s string, p ...string) string {
	var command string
	switch runtime.GOOS {
	case "windows":
		if !strings.HasSuffix(s, ".exe") {
			command = s + ".exe"
		}
	default:
		command = strings.TrimSuffix(s, ".exe")
	}
	params := strings.Join(p, " ")
	fullCmd := fmt.Sprintf("%s %s", command, params)
	return strings.Trim(fullCmd, " ")
}
