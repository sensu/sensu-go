package cmd

import (
	"bufio"
	"fmt"
	"os"
	"sync"
)

func copyLines(mu *sync.Mutex, in *os.File, out *os.File) {
	scanner := bufio.NewScanner(in)
	for scanner.Scan() {
		mu.Lock()
		_, _ = out.Write(append(scanner.Bytes(), '\n'))
		_ = out.Sync()
		mu.Unlock()
	}
}

func pipeLogsToFile(path string) error {
	logFile, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("couldn't open log file: %s", err)
	}
	stdoutReader, stdoutWriter, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("couldn't create stdout pipe: %s", err)
	}
	stderrReader, stderrWriter, err := os.Pipe()
	if err != nil {
		return fmt.Errorf("couldn't create stderr pipe: %s", err)
	}
	os.Stdout = stdoutWriter
	os.Stderr = stderrWriter
	var mu sync.Mutex
	go copyLines(&mu, stdoutReader, logFile)
	go copyLines(&mu, stderrReader, logFile)

	return nil
}
