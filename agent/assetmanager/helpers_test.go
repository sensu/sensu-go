package assetmanager

import (
	"crypto/sha512"
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

func stringToSHA512(str string) string {
	h := sha512.New()
	_, _ = h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}
