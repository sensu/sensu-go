package helpers

import (
	"errors"
	"io"
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

// InputData returns the content of filename, if provided, or the standard
// input. An error is returned if no input data is provided or if the file could
// not be open
func InputData(filename string) (io.Reader, error) {
	if filename == "" {
		if err := DetectEmptyStdin(os.Stdin); err != nil {
			return nil, err
		}
		return os.Stdin, nil
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if stat.IsDir() {
		return nil, errors.New("directories not supported yet")
	}

	return file, nil
}
