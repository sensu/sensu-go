package version

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

var (
	// Version stores the version of the current build (e.g. 2.0.0)
	Version = "dev"

	// PreReleaseIdentifier stores the pre-release identifier of the current build (eg. beta-2)
	PreReleaseIdentifier string

	// BuildDate stores the timestamp of the build (e.g. 2017-07-31T13:11:15-0700)
	BuildDate string

	// BuildSHA stores the git sha of the build (e.g. 8673bed0a9705083987b9ecbbc1cc0758df13dd2)
	BuildSHA string
)

var (
	ErrNoBuildIteration = errors.New("Build iteration could not be found. If running locally you must set SENSU_BUILD_ITERATION.")
	TagParseError       = errors.New("A build iteration could not be parsed from the tag")
)

var (
	buildNumberRE       = regexp.MustCompile(`[0-9]+$`)
	prereleaseVersionRE = regexp.MustCompile(`.*\-.*\.([0-9]+)\-[0-9]+$`)
	versionRE           = regexp.MustCompile(`^[0-9]\.[0-9]\.[0-9]`)
)

type BuildType string

const (
	Nightly BuildType = "nightly"
	Alpha   BuildType = "alpha"
	Beta    BuildType = "beta"
	RC      BuildType = "rc"
	Stable  BuildType = "stable"
)

// Semver returns full semantic versioning compatible identifier.
// Format: VERSION-PRERELEASE+METADATA
func Semver() string {
	version := Version
	if PreReleaseIdentifier != "" {
		version = version + "-" + PreReleaseIdentifier
	}

	gitSHA := BuildSHA
	if len(gitSHA) > 7 {
		gitSHA = gitSHA[:7]
	}

	return version + "#" + gitSHA
}

// BuildTypeFromTag discovers the BuildType of the git tag.
func BuildTypeFromTag(tag string) BuildType {
	if tag == "" {
		return Nightly
	}
	// String matching gives us the type of build tag
	for _, bt := range []BuildType{Alpha, Beta, RC, Stable} {
		if strings.Contains(tag, string(bt)) {
			return bt
		}
	}
	// tag exists but does not contain any of the above, this is a stable build
	return Stable
}

// Iteration will output an iteration number based on what type of build the git
// sha represents and the ci platform it is running on.
func Iteration(tag string) (string, error) {
	bt := BuildTypeFromTag(tag)
	if bt == Nightly {
		if travis := os.Getenv("TRAVIS"); travis == "true" {
			return os.Getenv("TRAVIS_BUILD_NUMBER"), nil
		}
		if appveyor := os.Getenv("APPVEYOR"); appveyor == "true" {
			return os.Getenv("APPVEYOR_BUILD_NUMBER"), nil
		}
		if bi := os.Getenv("SENSU_BUILD_ITERATION"); bi != "" {
			return bi, nil
		}
		return "", ErrNoBuildIteration
	}
	bi := buildNumberRE.FindString(tag)
	var err error
	if bi == "" {
		err = TagParseError
	}
	return bi, err
}

// GetPrereleaseVersion will output the version of a prerelease from its tag
func GetPrereleaseVersion(tag string) (string, error) {
	bt := BuildTypeFromTag(tag)
	switch bt {
	case Alpha, Beta, RC:
		matches := prereleaseVersionRE.FindStringSubmatch(tag)
		var bt string
		if len(matches) > 1 {
			bt = matches[1]
		}
		var err error
		if bt == "" {
			err = fmt.Errorf("a prerelease version could not be parsed from %q", tag)
		}
		return bt, err
	default:
		return "", fmt.Errorf("build type not supported for prerelease: %q", bt)
	}
}

// GetVersion will output the version of the build (without iteration)
func GetVersion(tag string) (string, error) {
	bt := BuildTypeFromTag(tag)
	baseVersion := versionRE.FindString(tag)
	if baseVersion == "" {
		baseVersion = "dev"
	}
	switch bt {
	case Nightly:
		return fmt.Sprintf("%s-%s", baseVersion, bt), nil
	case Alpha, Beta, RC:
		if baseVersion == "" {
			return "", fmt.Errorf("invalid tag: %q", tag)
		}
		pre, err := GetPrereleaseVersion(tag)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s-%s.%s", baseVersion, bt, pre), nil
	case Stable:
		if baseVersion == "" {
			return "", fmt.Errorf("invalid tag: %q", tag)
		}
		return baseVersion, nil
	default:
		panic("unreachable")
	}
}

// FullVersion will output the version of the build (with iteration)
func FullVersion(tag string) (string, error) {
	it, err := Iteration(tag)
	if err != nil {
		return "", err
	}
	ver, err := GetVersion(tag)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s", ver, it), nil
}
