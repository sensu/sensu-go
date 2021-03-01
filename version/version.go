package version

import (
	"fmt"
	"runtime"
	"runtime/debug"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	CommunityEditionSuffix  = "ce"
	EnterpriseEditionSuffix = "ee"
	InvalidEditionSuffix    = "invalid"
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

	// Edition stores the edition of the build
	// (e.g. community or enterprise)
	Edition string = "community"

	// GoVersion stores the version of Go used to build the binary
	// (e.g. go1.14.2)
	GoVersion string = runtime.Version()

	promBuildInfo = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "sensu_go_build_info",
			Help: "Sensu Go build information",
		},
		[]string{"version", "buildsha", "goversion"},
	).WithLabelValues(Version, BuildSHA, GoVersion)
)

func init() {
	prometheus.MustRegister(promBuildInfo)
	promBuildInfo.Set(1)

	// If we don't have a version because it has been manually built from source,
	// use Go build info to display the main module version
	if Version == "" {
		if buildInfo, ok := debug.ReadBuildInfo(); ok {
			Version = buildInfo.Main.Version
		}
	}
}

// Semver returns full semantic versioning compatible identifier.
// Format: VERSION-PRERELEASE+METADATA
func Semver() string {
	return Version
}

func SemverWithEditionSuffix() string {
	var editionSuffix string
	switch Edition {
	case "community":
		editionSuffix = CommunityEditionSuffix
	case "enterprise":
		editionSuffix = EnterpriseEditionSuffix
	default:
		editionSuffix = InvalidEditionSuffix
	}
	return fmt.Sprintf("%s+%s", Semver(), editionSuffix)
}

func EditionOutput() string {
	if Edition == "community" || Edition == "enterprise" {
		return fmt.Sprintf("%s edition", Edition)
	}
	return "built with an invalid \"edition\" ldflag"
}

func FormattedOutput(component string) string {
	var output string
	if component != "" {
		output += fmt.Sprintf("%s ", component)
	}
	output += fmt.Sprintf("version %s", SemverWithEditionSuffix())
	output += fmt.Sprintf(", %s", EditionOutput())
	if BuildSHA != "" {
		output += fmt.Sprintf(", build %s", BuildSHA)
	}
	if BuildDate != "" {
		output += fmt.Sprintf(", built %s", BuildDate)
	}
	output += fmt.Sprintf(", built with %s", GoVersion)
	return output
}

// Println prints all available details about the current version in a
// human-readable format, with an optional component name as the prefix
func Println(component string) {
	fmt.Println(FormattedOutput(component))
}
