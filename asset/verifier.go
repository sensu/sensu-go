package asset

import "os"

// A Verifier verifies that a file's SHA-512 matches the specified
// SHA-512.
type Verifier interface {
	Verify(file *os.File, sha512 string) error
}
