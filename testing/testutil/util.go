package testutil

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"runtime"
	"strings"
	"testing"
)

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

// RandomPorts generates len(p) random ports and assigns them to elements of p.
func RandomPorts(p []int) (err error) {
	for i := range p {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return err
		}
		defer func() {
			e := l.Close()
			if err == nil {
				err = e
			}
		}()

		addr, err := net.ResolveTCPAddr("tcp", l.Addr().String())
		if err != nil {
			return err
		}
		p[i] = addr.Port
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
