package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/sensu/sensu-go/internal/registry"
)

var (
	packagePath = flag.String("pkg", "", "Path to package to generate registry for")
	outPath     = flag.String("o", "", "Output package path")
)

func main() {
	log.SetFlags(log.Lshortfile)
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%s: Generate a type registry for sensu-go API types.\n", os.Args[0])
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Example usage: gen-register -pkg github.com/sensu/sensu-go -o register.go")
		flag.PrintDefaults()
	}
	flag.Parse()
	if *packagePath == "" {
		log.Fatal("no package path supplied (-pkg)")
	}
	if *outPath == "" {
		log.Fatal("no output path supplied (-o)")
	}
	if err := registry.RegisterTypes(*packagePath, *outPath); err != nil {
		log.Fatal(err)
	}
}
