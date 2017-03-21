package util

import (
	"io/ioutil"
	"log"
	"net"
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

// RandomPorts generates len(p) random ports and assigns them to elements of p.
func RandomPorts(p []int) error {
	for i := range p {
		l, err := net.Listen("tcp", ":0")
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
