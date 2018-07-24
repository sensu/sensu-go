package helpers

import (
	"errors"
	"os"
)

// DetectEmptyStdin determines if stdin is empty
func DetectEmptyStdin(f *os.File) error {
	fi, err := f.Stat()
	if err != nil {
		return err
	}
	if fi.Size() == 0 {
		if fi.Mode()&os.ModeNamedPipe == 0 {
			return errors.New("empty stdin")
		}
	}
	return nil
}
