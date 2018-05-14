# proto

[![Build Status](https://travis-ci.org/emicklei/proto.png)](https://travis-ci.org/emicklei/proto)
[![Go Report Card](https://goreportcard.com/badge/github.com/emicklei/proto)](https://goreportcard.com/report/github.com/emicklei/proto)
[![GoDoc](https://godoc.org/github.com/emicklei/proto?status.svg)](https://godoc.org/github.com/emicklei/proto)

Package in Go for parsing Google Protocol Buffers [.proto files version 2 + 3] (https://developers.google.com/protocol-buffers/docs/reference/proto3-spec)

This repository also includes 3 commands. The `protofmt` tool is for formatting .proto files. The `proto2xsd` tool is for generating XSD files from .proto version 3 files. The `proto2gql` tool is for generating the GraphQL Schema.

### usage as package

	package main

	import (
		"os"

		"github.com/emicklei/proto"
	)

	func main() {
		reader, _ := os.Open("test.proto")
		defer reader.Close()
		parser := proto.NewParser(reader)
		definition, _ := parser.Parse()
		formatter := proto.NewFormatter(os.Stdout, " ")
		formatter.Format(definition)
	}

### usage of proto2xsd command

	> proto2xsd -help
		Usage of proto2xsd [flags] [path ...]
  		-ns string
    		namespace of the target types (default "http://your.company.com/domain/version")		
  		-w	write result to an XSD files instead of stdout

See folder `cmd/proto2xsd/README.md` for more details.

### usage of proto2gql command

	> proto2gql -help
	    Usage of proto2gql [flags] [path ...]

        -std_out
            Writes transformed files to stdout
        -txt_out string
            Writes transformed files to .graphql file
        -go_out string
            Writes transformed files to .go file
        -js_out string
            Writes transformed files to .js file
        -package_alias value
            Renames packages using given aliases
        -resolve_import value
            Resolves given external packages
        -no_prefix
            Disables package prefix for type names

See folder `cmd/proto2gql/README.md` for more details.

### usage of protofmt command

	> protofmt -help
		Usage of protofmt [flags] [path ...]
  		-w	write result to (source) files instead of stdout

See folder `cmd/protofmt/README.md` for more details.

### how to install

    go get -u -v github.com/emicklei/proto

#### known issues

- the proto2 test file in protofmt folder contains character escape sequences that are currently not accepted by the scanner. See line 537 and 573.

© 2017, [ernestmicklei.com](http://ernestmicklei.com).  MIT License. Contributions welcome.