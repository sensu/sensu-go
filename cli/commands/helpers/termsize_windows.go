package helpers

import (
	"github.com/Azure/go-ansiterm/winterm"
)

// TermSize represents the size of a terminal.
type TermSize struct {
	Rows    uint16 // Rows is the terminal height, in rows.
	Columns uint16 // Columns is the terminal width, in columns.
}

// GetTermSize returns the terminal size based on the specified file descriptor.
func GetWinsize(fd uintptr) (*TermSize, error) {
	info, err := winterm.GetConsoleScreenBufferInfo(fd)
	if err != nil {
		return nil, err
	}

	ts := &TermSize{
		Columns: uint16(info.Window.Right - info.Window.Left + 1),
		Rows:    uint16(info.Window.Bottom - info.Window.Top + 1),
	}

	return ts, nil
}
