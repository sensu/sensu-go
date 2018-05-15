package main

import (
	"os"
	"io/ioutil"
	"path/filepath"
)

type FileWriter struct {
	filename string
	file     *os.File
	closeTag string
	isClosed bool
}

func NewFileWriter(filename string, openTag, closeTag string) (*FileWriter, error) {
	file, err := ioutil.TempFile(filepath.Dir(filename), "proto2gql")

	if err != nil {
		return nil, err
	}

	file.Write([]byte(openTag))

	return &FileWriter{filename, file, closeTag, false}, nil
}

func (fw *FileWriter) IsClosed() bool {
	return fw.isClosed
}

func (fw *FileWriter) Write(p []byte) (n int, err error) {
	return fw.file.Write(p)
}

func (fw *FileWriter) Save() error {
	if fw.isClosed == true {
		return nil // return an error?
	}

	tmpName := fw.file.Name()

	fw.file.Write([]byte(fw.closeTag))

	if err := fw.Close(); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(fw.filename), os.ModePerm); err != nil {
		return err
	}

	return os.Rename(tmpName, fw.filename)
}

func (fw *FileWriter) Remove() error {
	if fw.isClosed == true {
		return nil // return an error?
	}

	tmpName := fw.file.Name()

	if err := fw.Close(); err != nil {
		return err
	}

	return os.Remove(tmpName)
}

func (fw *FileWriter) Close() error {
	if fw.isClosed == false {
		fw.isClosed = true

		return fw.file.Close()
	}

	return nil
}