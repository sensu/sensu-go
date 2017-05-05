package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"syscall"
)

// processReader takes data from the reader and outputs to stdout.
// Based off the cat command in mc (https://github.com/minio/mc).
func processReader(r io.Reader) error {
	if _, err := io.Copy(os.Stdout, r); err != nil {
		switch err := err.(type) {
		case *os.PathError:
			if err.Err == syscall.EPIPE {
				// stdout closed by the user.
				// Gracefully exit.
				return nil
			}
			return err
		default:
			return err
		}
	}

	return nil
}

func readFile(f string) error {
	fileReader, err := os.Open(f)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	return processReader(fileReader)
}

func usage() {
	fmt.Println("usage: cat [file]")
}

func main() {
	argsCount := len(os.Args)
	switch argsCount {
	case 1:
		// no arguments provided, use stdin
		if err := processReader(os.Stdin); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	case 2:
		// argument specified, use argument as file path
		if err := readFile(os.Args[1]); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	default:
		usage()
	}
}
