package asset

import "os"

// resetFile ensures file contents are synced and rewound
func resetFile(f *os.File) error {
	if err := f.Sync(); err != nil {
		return err
	}
	_, err := f.Seek(0, 0)
	return err
}
