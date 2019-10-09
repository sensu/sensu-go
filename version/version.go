package version

import (
	"fmt"
	"runtime/debug"
)

// These values are baked into the sensu-agent and sensu-backend binaries
// during compilation, and are accessed using the Semver() function
var (
	// Version stores the version of the current build (e.g. 2.0.0)
	Version = ""

	// BuildDate stores the timestamp of the build
	// (e.g. 2017-07-31T13:11:15-0700)
	BuildDate string

	// BuildSHA stores the git sha of the build
	// (e.g. 8673bed0a9705083987b9ecbbc1cc0758df13dd2)
	BuildSHA string
)

// Semver returns full semantic versioning compatible identifier.
// Format: VERSION-PRERELEASE+METADATA
func Semver() string {
	version := Version

	// If we don't have a version because it has been manually built from source,
	// use Go build info to display the main module version
	if version == "" {
		buildInfo, ok := debug.ReadBuildInfo()
		if ok {
			version = buildInfo.Main.Version
		}
	}

	return version
}

// Println prints all available details about the current version in a
// human-readable format, with an optional component name as the prefix
func Println(component string) {
	var output string
	if component != "" {
		output += fmt.Sprintf("%s ", component)
	}
	output += fmt.Sprintf("version %s", Semver())
	if BuildSHA != "" {
		output += fmt.Sprintf(", build %s", BuildSHA)
	}
	if BuildDate != "" {
		output += fmt.Sprintf(", built %s", BuildDate)
	}

	fmt.Println(output)
}
