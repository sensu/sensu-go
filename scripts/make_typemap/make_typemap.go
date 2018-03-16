package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"
	"os/exec"
	"regexp"
)

var (
	tmplPath = flag.String("t", "", "Path to template file")
	output   = flag.String("o", "", "Path to output file")
	typeRe   = regexp.MustCompile(`^type ([A-Z].+) struct \{`)
)

type typeNames struct {
	TypeNames []string
}

func main() {
	flag.Parse()
	tmpl, err := template.ParseFiles(*tmplPath)
	if err != nil {
		log.Fatal(err)
	}
	typeNames, err := discoverTypeNames()
	if err != nil {
		log.Fatal(err)
	}
	out, err := os.Create(*output)
	if err != nil {
		log.Fatal(err)
	}
	if err := tmpl.Execute(out, typeNames); err != nil {
		log.Fatal(err)
	}
}

func discoverTypeNames() (typeNames, error) {
	var t typeNames
	doc, err := exec.Command("godoc", ".").Output()
	if err != nil {
		return t, err
	}
	scanner := bufio.NewScanner(bytes.NewReader(doc))
	for scanner.Scan() {
		line := scanner.Bytes()
		matches := typeRe.FindSubmatch(line)
		if len(matches) > 1 {
			fmt.Println(string(matches[1]))
			// capturing group match in matches[1]
			t.TypeNames = append(t.TypeNames, string(matches[1]))
		}
	}
	return t, scanner.Err()
}
