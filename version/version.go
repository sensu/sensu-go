package version

import "fmt"

var (
	Version   string
	Iteration string
	BuildDate string
	BuildSHA  string
)

// VersionWithIteration returns the version with iteration as a string
func VersionWithIteration() string {
	return fmt.Sprintf("%s-%s",
		Version,
		Iteration,
	)
}
