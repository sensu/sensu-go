// +build ignore

package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/mgutz/ansi"
	"github.com/sensu/sensu-go/backend/dashboardd/asset"
	"github.com/shurcooL/vfsgen"
)

const packageURL = "https://s3.us-west-2.amazonaws.com/sensu-ci-web-builds"
const packagePathPrefix = "oss/webapp"
const packageFilename = "dashboard.tgz"
const packageBranch = "embedded"

const filenamePrefix = "assets_"

// The ref file stores the git SHA associated with the embedded assets. It is
// used to short-curcuit needless updates of the asset box.
const refFile = ".ref"

var (
	red    = ansi.ColorFunc("red+b")
	white  = ansi.ColorFunc("white+b")
	yellow = ansi.ColorFunc("yellow+b")
)

func main() {
	if _, err := exec.LookPath("tar"); err != nil {
		printErr("'tar' was not found in your PATH, unable to box the web UI.")
		return
	}

	// pull latest ref from bucket
	ref, err := fetchLatestRef()
	if err != nil {
		printErr(err.Error())
		return
	}

	// read the git SHA associated with current state of the asset box
	lastRef, err := readLastRef()
	if lastRef == ref {
		fmt.Println("assets are already up to date")
		return
	}

	// download package
	fpath, err := fetchPackage(ref)
	if err != nil {
		printErr(err.Error())
		return
	}
	defer os.Remove(fpath)

	// extract package
	dir, err := extractPackage(fpath)
	if err != nil {
		printErr(err.Error())
		return
	}
	defer os.RemoveAll(dir)

	// combine bundled assets
	collection := asset.NewCollection()
	collection.Extend(http.Dir(filepath.Join(dir, "build", "app", "public")))
	collection.Extend(http.Dir(filepath.Join(dir, "build", "lib", "public")))
	collection.Extend(http.Dir(filepath.Join(dir, "build", "vendor", "public")))

	// box bundled assets
	err = vfsgen.Generate(collection, vfsgen.Options{
		Filename:     filenamePrefix + "oss" + ".go",
		PackageName:  "dashboard",
		VariableName: "OSS",
	})
	if err != nil {
		printErr(err.Error())
	}

	// save git SHA associated with the current state of the asset box
	if err = writeLastRef(ref); err != nil {
		printErr(err.Error())
	}
}

func getLastRefPath() string {
	_, currentFilePath, _, _ := runtime.Caller(0)
	dir := path.Dir(currentFilePath)
	return path.Join(dir, refFile)
}

func readLastRef() (string, error) {
	path := getLastRefPath()
	bs, err := ioutil.ReadFile(path)
	if err != nil && err != os.ErrNotExist {
		return "", err
	}
	return string(bs), nil
}

func writeLastRef(ref string) error {
	path := getLastRefPath()
	return ioutil.WriteFile(path, []byte(ref), 0644)
}

func fetchLatestRef() (string, error) {
	url := packageURL + "/" + path.Join(packagePathPrefix, packageBranch)
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	ref, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	if len(ref) != 40 {
		return "", fmt.Errorf("ref appears invalid, ref: %s...", ref[:48])
	}
	return string(ref), nil
}

func fetchPackage(ref string) (string, error) {
	tmpFile, err := ioutil.TempFile("", "*."+packageFilename)
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	url := packageURL + "/" + path.Join(packagePathPrefix, ref, packageFilename)
	res, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	_, err = io.Copy(tmpFile, res.Body)
	if err != nil {
		return "", err
	}

	if err = tmpFile.Sync(); err != nil {
		return "", err
	}
	return tmpFile.Name(), nil
}

func extractPackage(path string) (string, error) {
	tmpDir, err := ioutil.TempDir("", "")
	if err != nil {
		return "", err
	}
	mustRunCmd("tar", "-zxpf", path, "-C", tmpDir, "--strip-components=1")
	return tmpDir, nil
}

func printErr(msg string) {
	fmt.Println(red("⚠️  Error"), white(msg))
}

func mustRunCmd(pro string, args ...string) {
	cmd := exec.Command(pro, args...)
	cmdStr := strings.Join(append([]string{pro}, args...), " ")

	fmt.Printf("running '%s'\n", cmdStr)
	if buf, err := cmd.CombinedOutput(); err != nil {
		io.Copy(os.Stderr, bytes.NewReader(buf))
		fmt.Println("")
		fmt.Fprintf(os.Stderr, "🛑  %s %s '%s'\n", red("Error"), "failed to run", white(cmdStr))
		os.Exit(1)
	}
}
