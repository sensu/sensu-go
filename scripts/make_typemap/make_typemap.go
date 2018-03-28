package main

import (
	"bufio"
	"bytes"
	"flag"
	"html/template"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

var (
	tmplPath = flag.String("t", "", "Path to template file")
	output   = flag.String("o", "", "Path to output file")
	typeRe   = regexp.MustCompile(`^type ([A-Z].+) struct \{`)
)

type typeNames struct {
	TypeNames []string
}

func snakeCase(camelCase string) string {
	result := make([]rune, 0)
	for i, s := range camelCase {
		tl := strings.ToLower(string(s))
		if i == 0 {
			result = append(result, []rune(tl)...)
			continue
		}
		if string(s) == tl {
			result = append(result, s)
			continue
		}
		result = append(result, '_')
		result = append(result, []rune(tl)...)
	}
	return string(result)
}

func main() {
	flag.Parse()
	tmpl, err := template.New("typemap.tmpl").Funcs(template.FuncMap{
		"snakeCase": snakeCase,
	}).ParseFiles(*tmplPath)

	if err != nil {
		log.Fatalf("fatal error parsing typemap.tmpl: %s", err)
	}
	typeNames, err := discoverTypeNames()
	if err != nil {
		log.Fatalf("fatal error discovering types: %s", err)
	}
	out, err := os.Create(*output)
	if err != nil {
		log.Fatalf("fatal error creating typemap.go: %s", err)
	}
	if err := tmpl.Execute(out, typeNames); err != nil {
		log.Fatalf("fatal error generating typemap.go: %s", err)
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
			// capturing group match in matches[1]
			t.TypeNames = append(t.TypeNames, string(matches[1]))
		}
	}
	return t, scanner.Err()
}
