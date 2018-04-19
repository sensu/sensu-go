package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/vektra/mockery/mockery"
)

const (
	defaultOutDir = "./mocks"
	defaultOutPkg = "mocks"
)

type config struct {
	interfaces []string
	outDir     string
	outPkg     string
}

func main() {
	cfg := parseConfigFromArgs(os.Args)

	if len(cfg.interfaces) < 0 {
		fmt.Fprintln(os.Stderr, "must specify at least one interface to mock")
		os.Exit(1)
	}

	filter := regexp.MustCompile(
		fmt.Sprintf(
			"^(%s)$",
			strings.Join(cfg.interfaces, "|"),
		),
	)

	visitor := &mockery.GeneratorVisitor{
		Osp: &mockery.FileOutputStreamProvider{
			BaseDir: cfg.outDir,
			Case:    "underscore",
		},
		PackageName: cfg.outPkg,
	}
	walker := mockery.Walker{
		BaseDir: ".",
		Filter:  filter,
	}

	ok := walker.Walk(visitor)
	if !ok {
		fmt.Println("unable to find matching interfaces in this path")
		os.Exit(1)
	}
}

func parseConfigFromArgs(args []string) config {
	cfg := config{}

	flagSet := flag.NewFlagSet(args[0], flag.ExitOnError)
	flagSet.StringVar(&cfg.outDir, "outdir", defaultOutDir, "directory where to write mocks to")
	flagSet.StringVar(&cfg.outPkg, "outpkg", defaultOutPkg, "name of the generated package")
	flagSet.Usage = func() {
		_, err := fmt.Fprint(os.Stderr, "Usage: make_mock [OPTIONS] INTERFACE\n\n")
		if err != nil {
			fmt.Println("unable to write usage")
			os.Exit(1)
		}
		flagSet.PrintDefaults()
	}

	flagSet.Parse(args[1:])
	cfg.interfaces = flagSet.Args()
	return cfg
}
