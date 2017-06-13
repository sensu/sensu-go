package assetmanager

import (
	"crypto/sha256"
	"encoding/hex"
	"io/ioutil"
	"path/filepath"
	"runtime"
)

func readFixture(file string) string {
	_, curFilename, _, _ := runtime.Caller(0)
	curPath := filepath.Dir(curFilename)
	bytes, _ := ioutil.ReadFile(filepath.Join(curPath, "fixtures", file))
	return string(bytes)
}

func stringToSHA256(str string) string {
	h := sha256.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
