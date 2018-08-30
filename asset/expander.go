package asset

import "os"

// An Expander expands the provided *os.File to the target direcrtory.
type Expander interface {
	Expand(archive *os.File, targetDirectory string) error
}
