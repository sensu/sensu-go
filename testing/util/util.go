package util

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"runtime"
	"strings"
)

// WithTempDir runs function f within a temporary directory whose contents
// will be removed when execution of the function is finished.
func WithTempDir(f func(string)) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "sensu")
	defer os.RemoveAll(tmpDir)
	if err != nil {
		log.Panic(err)
	}
	f(tmpDir)
}

// RandomPorts generates len(p) random ports and assigns them to elements of p.
func RandomPorts(p []int) error {
	for i := range p {
		l, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			return err
		}
		defer l.Close()

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
