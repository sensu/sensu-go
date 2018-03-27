package testutil

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
)

// dirty hack to prevent port collisions
var lastPort int64 = 30000

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

// ReservePort reserves a port, as long as this process is the only one
// on the system opening ports. Introduced to avoid port collisions in
// our end-to-end tests. A total hack, and will probably need replacing
// one day.
func ReservePort() (int, error) {
	for {
		port := atomic.AddInt64(&lastPort, 1)
		if port > 65535 {
			return 0, errors.New("port allocation failed")
		}
		ln, err := net.Listen("tcp4", fmt.Sprintf("127.0.0.1:%d", port))
		if err != nil {
			if oe, ok := err.(*net.OpError); ok {
				if oe.Timeout() {
					return 0, err
				}
			}
			continue
		}
		for {
			err, ok := ln.Close().(*net.OpError)
			if err == nil {
				return int(port), nil
			}
			if ok && err.Temporary() {
				continue
			}
			return 0, err
		}
	}
}

// RandomPorts reserves len(p) ports and assigns them to elements of p.
// It calls ReservePort repeatedly.
func RandomPorts(p []int) (err error) {
	for i := range p {
		port, err := ReservePort()
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
