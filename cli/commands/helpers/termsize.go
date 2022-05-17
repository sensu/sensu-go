// +build !windows

package helpers

import (
	"golang.org/x/sys/unix"
)

// TermSize represents the size of a terminal.
type TermSize struct {
	Rows    uint16 // Rows is the terminal height, in rows.
	Columns uint16 // Columns is the terminal width, in columns.
}

// GetTermSize returns the terminal size based on the specified file descriptor.
func GetTermSize(fd uintptr) (*TermSize, error) {
	uws, err := unix.IoctlGetWinsize(int(fd), unix.TIOCGWINSZ)
	ts := &TermSize{Rows: uws.Row, Columns: uws.Col}
	return ts, err
}
