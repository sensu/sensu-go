package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/spf13/pflag"
)

var (
	// BuildDate is the date when the binary was built.
	BuildDate string
	// GitCommit is the commit hash that built the binary.
	GitCommit string
	// Version is the version.
	Version string
)

func main() {
	fmt.Printf("Version: %s - Commit: %s - Date: %s\n", Version, GitCommit, BuildDate)
	rand.Seed(time.Now().UnixNano())
	c := newCluster()
	c.AddFlags(pflag.CommandLine)
	pflag.Parse()
	c.Run()
}
