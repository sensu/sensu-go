package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	fromPath = flag.String("from", "", "Package to freeze")
	toPath   = flag.String("to", "", "Versioned package to create")
)

func main() {
	flag.Parse()
	if *fromPath == "" || *toPath == "" {
		flag.Usage()
		os.Exit(1)
	}
	if err := freezeAPI(*fromPath, *toPath); err != nil {
		log.Fatal(err)
	}
}

func freezeAPI(from, to string) error {
	if err := copyPackage(from, to); err != nil {
		return fmt.Errorf("error copying packages: %s", err)
	}
	if err := createConverters(from, to); err != nil {
		return fmt.Errorf("error creating converters: %s", err)
	}
	if err := registerTypes(to); err != nil {
		return fmt.Errorf("error registering types: %s", err)
	}
	return nil
}

func createConverters(from, to string) error {
	return nil
}

func registerTypes(to string) error {
	return nil
}
