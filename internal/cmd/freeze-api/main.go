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
	successMessage := fmt.Sprintf("Froze API to %q. Now run: 'go generate github.com/sensu/sensu-go/runtime/registry'.", *toPath)
	fmt.Fprintln(flag.CommandLine.Output(), successMessage)
}

func freezeAPI(from, to string) error {
	if err := copyPackage(from, to); err != nil {
		return fmt.Errorf("error copying packages: %s", err)
	}
	if err := createConverters(to, from); err != nil {
		return fmt.Errorf("error creating converters: %s", err)
	}
	return nil
}
