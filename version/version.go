package version

import "fmt"

var (
	// Version stores the version of the current build (e.g. 2.0.0)
	Version string

	// Iteration stores the iteration of the current iteration (e.g. 1)
	Iteration string

	// BuildDate stores the timestamp of the build (e.g. 2017-07-31T13:11:15-0700)
	BuildDate string

	// BuildSHA stores the git sha of the build (e.g. 8673bed0a9705083987b9ecbbc1cc0758df13dd2)
	BuildSHA string
)

// WithIteration returns the version with iteration as a string
func WithIteration() string {
	return fmt.Sprintf("%s-%s",
		Version,
		Iteration,
	)
}
