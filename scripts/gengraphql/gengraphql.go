package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"

	"github.com/sensu/sensu-go/graphql/generator"
)

const (
	packageNameDefault = "schema"
	debuggingDefault   = false
)

var logger = logrus.WithField("component", "scripts/gengraphql")

func main() { // nolint
	// Parse
	config := parseArgs(os.Args)

	// Configure logger
	if config.debug {
		logger.Logger.Level = logrus.DebugLevel
		logger.
			WithFields(logrus.Fields{
				"path":         config.path,
				"debug":        config.debug,
				"package-name": config.pkgName,
			}).
			Debug("configured")
	}

	// Find GraphQL files
	files, err := ioutil.ReadDir(config.path)
	if err != nil {
		logger.Fatal(err)
	}
	for i := 0; i < len(files); {
		if present := strings.HasSuffix(files[i].Name(), "graphql"); present {
			logger.WithField("file", files[i].Name()).Debug("graphql file found")
			i++
		} else {
			files = append(files[:i], files[i+1:]...)
		}
	}

	// Parse
	logger.Info("parsing .graphql files")
	graphqlFiles := make(generator.GraphQLFiles, len(files))
	for i, f := range files {
		log := logger.WithField("file", f.Name())
		filePath := filepath.Join(config.path, f.Name())

		file, err := generator.ParseFile(filePath)
		if err != nil {
			log.WithError(err).Fatal("unable to parse file; check syntax")
		}

		if err := file.Validate(); err != nil {
			log.Fatal(err)
		}

		log.Debug("file parsed successfully")
		graphqlFiles[i] = file
	}

	// Generate
	generator := generator.New(graphqlFiles)
	generator.PackageName = config.pkgName
	generator.Invoker = "scripts/gengraphql.go"

	if err := generator.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		logger.Fatal("unable to generate file")
	}
	logger.Info("files written successfully")
}

type config struct {
	path    string
	pkgName string
	debug   bool
}

func parseArgs(args []string) config {
	c := config{}

	cmd := flag.NewFlagSet(args[0], flag.ExitOnError)
	cmd.StringVar(&c.pkgName, "P", packageNameDefault, "name of the package")
	cmd.BoolVar(&c.debug, "debug", debuggingDefault, "print debug messages")
	cmd.Usage = func() {
		_, err := fmt.Fprint(os.Stderr, "Usage: gengraphql [OPTIONS] DIRECTORY\n\n")
		if err != nil {
			logger.Fatal("unable to write usage")
		}
		cmd.PrintDefaults()
	}

	if len(args) < 2 {
		cmd.Usage()
		os.Exit(1)
	}

	if err := cmd.Parse(args[1:]); err != nil {
		cmd.Usage()
		os.Exit(1)
	}

	// Ensure directory argument was given
	if cmd.NArg() < 1 {
		cmd.Usage()
		os.Exit(1)
	}

	c.path = cmd.Arg(0)
	return c
}
