// +build ignore

package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mgutz/ansi"
	"github.com/sensu/sensu-go/backend/dashboardd/asset"
	"github.com/shurcooL/vfsgen"
)

const filenamePrefix = "assets_"
const buildDir = "./build"

var (
	red    = ansi.ColorFunc("red+b")
	white  = ansi.ColorFunc("white+b")
	yellow = ansi.ColorFunc("yellow+b")
)

func main() {
	if _, err := exec.LookPath("node"); err != nil {
		fmt.Println(yellow("⚠️  Warning"), white("'node' was not found in your PATH, unable to bundle web UI."))
		fmt.Println(white("See https://nodejs.org/en/download/package-manager/ for installation instructions."))
		fmt.Println(white("Skipping dashboard build."))
		return
	}

	if _, err := exec.LookPath("yarn"); err != nil {
		fmt.Println(yellow("⚠️  Warning"), white("'yarn' was not found in your PATH, unable to bundle web UI."))
		fmt.Println(white("See https://yarnpkg.com/en/docs/install for installation instructions."))
		fmt.Println(white("Skipping dashboard build."))
		return
	}

	// install web ui depedencies
	mustRunCmd("yarn", "install")

	// install web ui depedencies
	mustRunCmd("yarn", "build")

	// combine bundled assets
	collection := asset.NewCollection()
	collection.Extend(http.Dir(filepath.Join(buildDir, "app", "public")))
	collection.Extend(http.Dir(filepath.Join(buildDir, "lib", "public")))
	collection.Extend(http.Dir(filepath.Join(buildDir, "vendor", "public")))

	// box bundled assets
	err := vfsgen.Generate(collection, vfsgen.Options{
		Filename:     filenamePrefix + "oss" + ".go",
		PackageName:  "dashboard",
		VariableName: "OSS",
	})
	if err != nil {
		log.Fatalln(err)
	}
}

func mustRunCmd(pro string, args ...string) {
	cmd := exec.Command(pro, args...)
	cmd.Env = append(os.Environ(), "NODE_ENV=production")
	cmdStr := strings.Join(append([]string{pro}, args...), " ")

	fmt.Printf("Running '%s'\n", cmdStr)
	if buf, err := cmd.CombinedOutput(); err != nil {
		fmt.Println("")
		io.Copy(os.Stderr, bytes.NewReader(buf))
		fmt.Fprintf(os.Stderr, "🛑  %s %s '%s'\n", red("Error"), "failed to run", white(cmdStr))
		os.Exit(1)
	}
}
