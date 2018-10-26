package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
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
	if err := createConverters(from, to); err != nil {
		return fmt.Errorf("error freezing API: %s", err)
	}
	if err := createResourceNameMethods(from, to); err != nil {
		return fmt.Errorf("error freezing API: %s", err)
	}
	if out, err := exec.Command("go", "generate", registryPath).CombinedOutput(); err != nil {
		return fmt.Errorf("error freezing API: %s", string(out))
	}
	if out, err := exec.Command("go", "fmt", "github.com/sensu/sensu-go/...").CombinedOutput(); err != nil {
		return fmt.Errorf("error freezing API: %s", out)
	}
	if out, err := exec.Command("./hack/update-protobuf.sh").CombinedOutput(); err != nil {
		return fmt.Errorf("error freezing API: %s", out)
	}

	return nil
}
