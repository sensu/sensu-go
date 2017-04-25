package test

import (
	"io/ioutil"
	"os"
)

// StdoutCapture overwrites Stdout to capture bytes written to it
// Borrowed from testify.suite package
// https://github.com/stretchr/testify/blob/v1.1.4/suite/suite_test.go
type StdoutCapture struct {
	oldStdout *os.File
	readPipe  *os.File

	Bytes string
}

// StartCapture overwrites Stdout and starts capturing
func (sc *StdoutCapture) StartCapture() {
	sc.Bytes = ""
	sc.oldStdout = os.Stdout
	sc.readPipe, os.Stdout, _ = os.Pipe()
}

// StopCapture stops capture and restores Stdout
func (sc *StdoutCapture) StopCapture() {
	if sc.oldStdout == nil || sc.readPipe == nil {
		panic("StartCapture not called before StopCapture")
	}
	os.Stdout.Close()
	os.Stdout = sc.oldStdout
	bytes, err := ioutil.ReadAll(sc.readPipe)
	if err == nil {
		sc.Bytes = string(bytes)
	}
}
