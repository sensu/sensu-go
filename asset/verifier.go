package asset

import (
	"crypto/sha512"
	"encoding/hex"
	"fmt"
	"io"
)

var (
	defaultVerifier = &sha512Verifier{}
)

// A Verifier verifies that a file's SHA-512 matches the specified
// SHA-512.
type Verifier interface {
	Verify(file io.ReadSeeker, sha512 string) error
}

// sha512Verifier verifies that a file matches a specified SHA-512 sum.
type sha512Verifier struct{}

// Verify that the file matches the desired SHA-512 sum.
func (v *sha512Verifier) Verify(rs io.ReadSeeker, desiredSHA string) error {
	// Generate checksum for downloaded file
	h := sha512.New()
	if _, err := io.Copy(h, rs); err != nil {
		return fmt.Errorf("generating checksum for asset failed: %s", err)
	}

	if _, err := rs.Seek(0, 0); err != nil {
		return err
	}

	if foundSHA := hex.EncodeToString(h.Sum(nil)); foundSHA != desiredSHA {
		return fmt.Errorf("sha512 '%s' does not match specified sha512: '%s'", desiredSHA, foundSHA)
	}

	return nil
}
