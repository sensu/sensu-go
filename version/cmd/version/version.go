package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/sensu/sensu-go/version"
)

var (
	fFullVersion = flag.Bool("f", false, "output the version of the build with iteration (ex: 2.0.0-alpha.17-1)")
	fBaseVersion = flag.Bool("b", false, "output the base version (ex: 2.0.1)")
	fIteration   = flag.Bool("i", false, "output the iteration of the build (ex: the 1 in 2.0.0-alpha.17-1)")
	fPrerelease  = flag.Bool("p", false, "output the prerelease version of the build (ex: 17 from tag 2.0.0-alpha.17)")
	fBuildType   = flag.Bool("t", false, "output the type of build this is (alpha|beta|rc|dev|nightly|stable)")
	fVersion     = flag.Bool("v", false, "output the version of the build without iteration (ex: 2.0.0-alpha.17)")
)

func main() {
	flag.Parse()
	tag, bt, err := version.FindVersionInfo(&BuildEnv{})
	if err != nil {
		log.Fatal(err)
	}
	var fn func(string, version.BuildType) (string, error)
	if *fFullVersion {
		fn = version.FullVersion
	} else if *fBaseVersion {
		fn = version.GetBaseVersion
	} else if *fIteration {
		fn = version.Iteration
	} else if *fPrerelease {
		fn = version.GetPrereleaseVersion
	} else if *fBuildType {
		fn = buildType
	} else if *fVersion {
		fn = version.GetVersion
	} else {
		flag.PrintDefaults()
		os.Exit(1)
	}
	result, err := fn(tag, bt)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(result)
}

// buildType getter func that matches the signature of other version funcs
func buildType(_ string, bt version.BuildType) (string, error) {
	return string(bt), nil
}

// implements the version.BuildEnv interface using the real build environment
type BuildEnv struct{}

// Returns true if we are building from a CI (rather than a local or dev build)
func (BuildEnv) IsCI() bool {
	// Travis, AppVeyor, CircleCI, and many others all set CI=true in their
	// environment by default.
	return os.Getenv("CI") == "true"
}

// Returns true if this is a nightly release by checking whether the current
// HEAD is one or more commits ahead of the latest tag.
func (BuildEnv) IsNightly() (bool, error) {
	cmd := exec.Command("git", "describe", "--exact-match", "--tags", "HEAD")
	err := cmd.Run()
	// if the tag is an exact match for current HEAD, this is not a nightly
	if err == nil {
		return false, nil
	}
	// if the command exited with a nonzero status, this is a nightly build
	if _, ok := err.(*exec.ExitError); ok {
		return true, nil
	}
	// if the command somehow failed to execute, it's an error
	return false, err
}

// Returns the most recent tag belonging to the current commit
func (BuildEnv) GetMostRecentTag() (string, error) {
	// --abbrev=0 disables tag annotation, returning unmodified most recent tag
	cmd := exec.Command("git", "describe", "--abbrev=0", "--tags", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	tag := strings.Trim(string(out), "\n")
	return tag, nil
}
