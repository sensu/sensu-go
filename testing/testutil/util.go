package testutil

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// TempDir provides a test with a temporary directory (under os.TempDir())
// returning the absolute path to the directory and a remove() function
// that should be deferred immediately after calling TempDir(t) to recursively
// delete the contents of the directory.
func TempDir(t testing.TB) (tmpDir string, remove func()) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "sensu")
	if err != nil {
		t.FailNow()
	}

	return tmpDir, func() { _ = os.RemoveAll(tmpDir) }
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

// RandomUUIDV4 takes a testing.TB and will attempt to generate a random
// Version 4 UUID. If an error is returned, a fatal testing error will occur.
func RandomUUIDV4(tb testing.TB) uuid.UUID {
	bytes, err := uuid.NewRandom()
	if err != nil {
		tb.Fatalf("failed to generate uuid: %s", err)
	}
	return bytes
}
