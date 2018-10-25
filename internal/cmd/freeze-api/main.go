package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
)

const registryPath = "github.com/sensu/sensu-go/runtime/registry"

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
	successMessage := fmt.Sprintf("Froze API to %q.", *toPath)
	fmt.Fprintln(flag.CommandLine.Output(), successMessage)
}

func freezeAPI(from, to string) error {
	if err := copyPackage(from, to); err != nil {
		return fmt.Errorf("error freezing API: %s", err)
	}
	if err := createConverters(to, from); err != nil {
		return fmt.Errorf("error freezing API: %s", err)
	}
	if err := createResourceNameMethods(from, to); err != nil {
		return fmt.Errorf("error freezing API: %s", err)
	}

	if err := exec.Command("go", "generate", registryPath).Run(); err != nil {
		return fmt.Errorf("error freezing API: %s", err)
	}

	if err := exec.Command("go", "fmt", path.Join(*toPath, "...")).Run(); err != nil {
		return fmt.Errorf("error freezing API: %s", err)
	}

	return nil
}
