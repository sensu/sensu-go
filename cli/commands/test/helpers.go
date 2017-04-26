package test

import (
	"io/ioutil"
	"os"
)

type FileCapture struct {
	file          **os.File
	originalValue *os.File
	reader        *os.File
	output        string
}

func NewFileCapture(f **os.File) FileCapture {
	return FileCapture{file: f}
}

func (fc *FileCapture) Start() {
	// copy the file into oldValue
	fc.originalValue = *fc.file

	// create a pipe returning reader and writer files, assign
	// reference to the writer file to the file pointer
	fc.reader, *fc.file, _ = os.Pipe()
}

func (fc *FileCapture) Stop() {
	// store reference to the writer file
	writer := *fc.file

	// reassign the reference to the original file back to the
	// file pointer
	*fc.file = fc.originalValue

	// close the writer file as it's no longer needed
	writer.Close()

	// store the contents of the reader as a string
	// in fc.output
	bytes, _ := ioutil.ReadAll(fc.reader)
	fc.output = string(bytes)
}

func (fc *FileCapture) Output() string {
	return fc.output
}
