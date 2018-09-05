package asset

import (
	"fmt"
	"io"

	"github.com/mholt/archiver"

	filetype "gopkg.in/h2non/filetype.v1"
	filetype_types "gopkg.in/h2non/filetype.v1/types"
)

const (
	// Size of file header for sniffing type
	headerSize = 262
)

// An expander expands the provided *os.File to the target direcrtory.
type expander interface {
	Expand(archive io.ReadSeeker, targetDirectory string) error
}

// A archiveExpander detects the archive type and expands it to the local
// filesystem.
//
// Supported archive types:
// - tar
// - tar-gzip
type archiveExpander struct{}

// Expand an archive to a target directory.
func (a *archiveExpander) Expand(archive io.ReadSeeker, targetDirectory string) error {
	// detect the type of archive the asset is
	ft, err := sniffType(archive)
	if err != nil {
		return err
	}

	var ar archiver.Archiver

	// If the file is not an archive, exit with an error.
	switch ft.MIME.Value {
	case "application/x-tar":
		ar = archiver.Tar
	case "application/gzip":
		ar = archiver.TarGz
	default:
		return fmt.Errorf(
			"given file of format '%s' does not appear valid",
			ft.MIME.Value,
		)
	}

	// Extract the archive to the desired path
	if err := ar.Read(archive, targetDirectory); err != nil {
		return fmt.Errorf("error extracting asset: %s", err)
	}

	return nil
}

func sniffType(f io.ReadSeeker) (filetype_types.Type, error) {
	header := make([]byte, headerSize)
	if _, err := f.Read(header); err != nil {
		return filetype_types.Type{}, fmt.Errorf("unable to read asset header: %s", err)
	}
	ft, err := filetype.Match(header)
	if err != nil {
		return ft, err
	}

	if _, err := f.Seek(io.SeekStart, io.SeekStart); err != nil {
		return filetype_types.Type{}, err
	}

	return ft, nil
}
