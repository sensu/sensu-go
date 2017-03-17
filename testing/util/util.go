package util

import (
	"io/ioutil"
	"log"
	"os"
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
