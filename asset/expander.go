package asset

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	filetype "gopkg.in/h2non/filetype.v1"
	filetype_types "gopkg.in/h2non/filetype.v1/types"
)

const (
	// Size of file header for sniffing type
	headerSize = 262
)

var (
	defaultExpander = &archiveExpander{}
)

// An Expander expands the provided *os.File to the target direcrtory.
type Expander interface {
	Expand(archive io.ReadSeeker, targetDirectory string) error
}

// A archiveExpander detects the archive type and expands it to the local
// filesystem.
//
// Supported archive types:
// - tar
// - tar-gzip
type archiveExpander struct{}

type namer interface {
	Name() string
}

// Expand an archive to a target directory.
func (a *archiveExpander) Expand(archive io.ReadSeeker, targetDirectory string) error {
	// detect the type of archive the asset is
	ft, err := sniffType(archive)
	if err != nil {
		return err
	}

	namer, ok := archive.(namer)
	if !ok {
		return errors.New("couldn't get path to archive")
	}

	// If the file is not an archive, exit with an error.
	switch ft.MIME.Value {
	case "application/x-tar":
		if err := extractTar(namer.Name(), targetDirectory); err != nil {
			return fmt.Errorf("error extracting asset: %s", err)
		}
	case "application/gzip":
		if err := extractTarGz(namer.Name(), targetDirectory); err != nil {
			return fmt.Errorf("error extracting asset: %s", err)
		}
	default:
		return fmt.Errorf(
			"given file of format '%s' does not appear valid",
			ft.MIME.Value,
		)
	}

	return nil
}

// extractTar extracts a tar archive at srcPath into destDir.
func extractTar(srcPath, destDir string) error {
	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return untar(f, destDir)
}

// extractTarGz extracts a gzip-compressed tar archive at srcPath into destDir.
func extractTarGz(srcPath, destDir string) error {
	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close()

	gr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gr.Close()

	return untar(gr, destDir)
}

// untar reads a tar stream and extracts files into destDir.
func untar(r io.Reader, destDir string) error {
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		// Sanitize the name to prevent path traversal.
		cleanName := filepath.Clean(hdr.Name)
		if strings.HasPrefix(cleanName, "..") {
			return fmt.Errorf("invalid file path in archive: %s", hdr.Name)
		}

		target := filepath.Join(destDir, cleanName)

		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(hdr.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			if err := os.Symlink(hdr.Linkname, target); err != nil {
				return err
			}
		}
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

	if _, err := f.Seek(0, 0); err != nil {
		return filetype_types.Type{}, err
	}

	return ft, nil
}
