package testing

import (
	"io/ioutil"
	"os"
)

// FileCapture helps us write tests where we want to assert that
// something was written to STDOUT or STDERR.
//
// Usage:
//
//   stdout := NewFileCapture(&os.Stdin)
//   stdout.Start()
//   fmt.Println("omgomgomg")
//   stdout.Stop()
//   assert.Equal("omgomgomg", stdout.Output())
//
type FileCapture struct {
	file          **os.File
	originalValue *os.File
	reader        *os.File
	output        string
}

// NewFileCapture instantiates new FileCapture
func NewFileCapture(f **os.File) FileCapture {
	return FileCapture{file: f}
}

// Start starts capturing data written to FileCapture#file
func (fc *FileCapture) Start() {
	// copy the file into oldValue
	fc.originalValue = *fc.file

	// create a pipe returning reader and writer files, assign
	// reference to the writer file to the file pointer
	fc.reader, *fc.file, _ = os.Pipe()
}

// Stop stops capturing and reads the data to the FileCaputre#output string
func (fc *FileCapture) Stop() {
	// store reference to the writer file
	writer := *fc.file

	// reassign the reference to the original file back to the
	// file pointer
	*fc.file = fc.originalValue

	// close the writer file as it's no longer needed
	_ = writer.Close()

	// store the contents of the reader as a string
	// in fc.output
	bytes, _ := ioutil.ReadAll(fc.reader)
	fc.output = string(bytes)
}

// Output exposes the data captured; only available after capture has stopped
func (fc *FileCapture) Output() string {
	return fc.output
}
