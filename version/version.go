package version

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// These values are baked into the sensu-agent and sensu-backend binaries
// during compilation, and are accessed using the Semver() function
var (
	// Version stores the version of the current build (e.g. 2.0.0)
	Version = ""

	// PreReleaseIdentifier stores the pre-release identifier of the current
	// build (e.g. the 2 in beta-2)
	PreReleaseIdentifier string

	// BuildDate stores the timestamp of the build
	// (e.g. 2017-07-31T13:11:15-0700)
	BuildDate string

	// BuildSHA stores the git sha of the build
	// (e.g. 8673bed0a9705083987b9ecbbc1cc0758df13dd2)
	BuildSHA string
)

var (
	ErrNoBuildIteration = errors.New("Build iteration could not be found. If running locally you must set SENSU_BUILD_ITERATION.")
	ErrTagParse         = errors.New("A build iteration could not be parsed from the tag")
)

var (
	buildNumberRE       = regexp.MustCompile(`[0-9]+$`)
	prereleaseVersionRE = regexp.MustCompile(`.*\-.*\.([0-9]+)\-[0-9]+$`)
	versionRE           = regexp.MustCompile(`^[0-9]\.[0-9]\.[0-9]`)
)

type BuildType string

const (
	Dev     BuildType = "dev"
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

// BuildEnv provides methods for determining version info from the current env.
type BuildEnv interface {
	IsCI() bool
	IsNightly() (bool, error)
	GetMostRecentTag() (string, error)
}

// FindVersionInfo discovers the most recent tag and BuildType using the
// current build environment
func FindVersionInfo(env BuildEnv) (string, BuildType, error) {
	tag, err := env.GetMostRecentTag()
	if err != nil {
		return "", "", err
	}
	// if building outside of CI, this is a dev build, regardless of tag
	if !env.IsCI() {
		return tag, Dev, nil
	}
	// if building from CI from a non-release commit, this is a nightly build
	isNightly, err := env.IsNightly()
	if err != nil {
		return "", "", err
	}
	if isNightly {
		return tag, Nightly, nil
	}
	// detect build type via string comparison on current tag
	for _, bt := range []BuildType{Alpha, Beta, RC, Stable} {
		if strings.Contains(tag, string(bt)) {
			return tag, bt, nil
		}
	}
	// tag doesn't match any other condition, default to stable build
	return tag, Stable, nil
}

// Iteration will output an iteration number based on what type of build the git
// sha represents and the ci platform it is running on.
// (e.g. the 1 in 2.0.0-alpha.17-1)
func Iteration(tag string, bt BuildType) (string, error) {
	if bt == Nightly {
		if bi := os.Getenv("SENSU_BUILD_ITERATION"); bi != "" {
			return bi, nil
		}
		return "", ErrNoBuildIteration
	}
	bi := buildNumberRE.FindString(tag)
	var err error
	if bi == "" {
		err = ErrTagParse
	}
	return bi, err
}

// GetPrereleaseVersion will output the version of a prerelease from its tag
// (e.g. "17" from tag "2.0.0-alpha.17")
func GetPrereleaseVersion(tag string, bt BuildType) (string, error) {
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
		// prerelease version does not apply to dev or nightly builds
		return "", nil
	}
}

// GetBaseVersion will output the major, minor, and patch #s with dots.
// (e.g. "2.0.1")
func GetBaseVersion(tag string, bt BuildType) (string, error) {
	baseVersion := versionRE.FindString(tag)
	if baseVersion == "" {
		return "", fmt.Errorf("Could not determine base version from %q", tag)
	}
	return baseVersion, nil
}

// GetVersion will output the version of the build (without iteration)
// (e.g. "2.0.0-alpha.17")
func GetVersion(tag string, bt BuildType) (string, error) {
	baseVersion, err := GetBaseVersion(tag, bt)
	if err != nil {
		return "", err
	}
	switch bt {
	case Dev, Nightly:
		return fmt.Sprintf("%s-%s", baseVersion, bt), nil
	case Alpha, Beta, RC:
		if baseVersion == "" {
			return "", fmt.Errorf("invalid tag: %q", tag)
		}
		pre, err := GetPrereleaseVersion(tag, bt)
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
// (e.g. "2.0.0-alpha.17-1")
func FullVersion(tag string, bt BuildType) (string, error) {
	it, err := Iteration(tag, bt)
	if err != nil {
		return "", err
	}
	ver, err := GetVersion(tag, bt)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s-%s", ver, it), nil
}
