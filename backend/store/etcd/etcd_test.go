package etcd

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEtcd(t *testing.T) {
	tmpDir, err := ioutil.TempDir(os.TempDir(), "sensu")
	if err != nil {
		assert.FailNow(t, err.Error())
	}
	defer os.RemoveAll(tmpDir)

	cfg := NewConfig()
	cfg.StateDir = tmpDir

	err = NewEtcd(cfg)
	assert.NoError(t, err)
}
