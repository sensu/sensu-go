package version

var (
	// Version stores the version of the current build (e.g. 2.0.0)
	Version string

	// PreReleaseIdentifier stores the pre-release identifier of the current build (eg. beta-2)
	PreReleaseIdentifier string

	// BuildDate stores the timestamp of the build (e.g. 2017-07-31T13:11:15-0700)
	BuildDate string

	// BuildSHA stores the git sha of the build (e.g. 8673bed0a9705083987b9ecbbc1cc0758df13dd2)
	BuildSHA string
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
