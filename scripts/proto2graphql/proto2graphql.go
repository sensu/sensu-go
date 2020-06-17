package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/emicklei/proto"
)

// Quick and dirty script that cobbles together a GraphQL definition from given
// protobuf file. Script is intented to only be run once for each message.
func main() { // nolint
	// script IN OUT
	if len(os.Args) < 3 {
		_, script := filepath.Split(os.Args[0])
		fmt.Printf("Usage: %s PROTOFILE SCHEMAPATH\n", script)
		os.Exit(1)
	}

	// Open file
	infile := os.Args[1]
	reader, _ := os.Open(infile)
	defer reader.Close()

	// Open outfile
	outfile := outfilepath(infile, os.Args[2])
	if _, err := os.Stat(outfile); !os.IsNotExist(err) {
		fmt.Println(outfile + " already exists. Not intented not be run more than once.")
		os.Exit(1)
	}

	// Parse
	parser := proto.NewParser(reader)
	definition, err := parser.Parse()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var out string
	for _, elem := range definition.Elements {
		message, ok := elem.(*proto.Message)
		if !ok {
			continue
		}

		//
		// TODO:
		//
		// Ultimately this should use graphql-go/graphql/language/printer, however,
		// it currently does not print descriptions. Whoops.
		//
		// Strap in this is going to get ugly.
		//

		// Write comment
		out += genComment(message.Comment, true, 0)

		// Write type
		out += fmt.Sprintf("type %s {\n", message.Name)

		// Write fields
		for _, node := range message.Elements {
			field, ok := node.(*proto.NormalField)
			if !ok {
				continue
			}

			name := underscoreToCamel(field.Name)
			var t string
			switch field.Type {
			case "bytes":
				fallthrough
			case "string":
				t = "String"
			case "uint64":
				fallthrough
			case "uint32":
				fallthrough
			case "int64":
				fallthrough
			case "int32":
				t = "Int"
			case "bool":
				t = "Boolean"
			default:
				t = field.Type
			}

			if field.Repeated {
				t = fmt.Sprintf("[%s!]", t)
			}
			if !isNullable(field.Field) {
				t = fmt.Sprintf("%s!", t)
			}

			out += genComment(field.Comment, false, 2)
			out += fmt.Sprintf("  %s: %s\n", name, t)
		}

		// Close type
		out += "}\n\n"
	}

	// Write to file
	err = ioutil.WriteFile(outfile, []byte(out), 0644)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Success message
	fmt.Printf("Wrote: %s\n", outfile)
}

func outfilepath(infile string, path string) string {
	_, fname := filepath.Split(infile)
	outfname := fname[:(len(fname) - len(filepath.Ext(fname)))]
	abspath, err := filepath.Abs(path)
	if err != nil {
		panic(err)
	}
	return filepath.Join(abspath, outfname+".graphql")
}

// https://github.com/asaskevich/govalidator/blob/3153c74/utils.go#L101
func underscoreToCamel(in string) string {
	head := in[:1]
	repl := strings.Replace(
		strings.Title(strings.Replace(strings.ToLower(in), "_", " ", -1)),
		" ",
		"",
		-1,
	)
	return head + repl[1:]
}

func genComment(comment *proto.Comment, forceMulti bool, indentLvl int) string {
	if comment == nil {
		return ""
	}

	indent := indentStr(indentLvl)
	if len(comment.Lines) > 1 || forceMulti {
		var out string
		for _, line := range comment.Lines {
			out += indent
			out += strings.TrimSpace(line)
			out += "\n"
		}
		return fmt.Sprintf("%s\"\"\"\n%s%s\"\"\"\n", indent, out, indent)
	}

	return fmt.Sprintf(
		"%s\"%s\"\n",
		indent,
		strings.TrimSpace(comment.Message()),
	)
}

func isNullable(field *proto.Field) bool {
	for _, opt := range field.Options {
		if opt.Name == "(gogoproto.nullable)" || opt.Constant.Source == "true" {
			return true
		}
	}
	return false
}

func indentStr(len int) string {
	var out string
	for i := 0; i < len; i++ {
		out += " "
	}
	return out
}
