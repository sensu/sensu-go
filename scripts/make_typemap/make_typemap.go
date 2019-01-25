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

		// Treat acronyms as single-word, e.g. LDAP -> ldap
		var nextCharCaseChanges bool
		if i+1 < len(camelCase) {
			nextChar := camelCase[i+1]
			// Check if the next character case differs from the current character
			if (s >= 'a' && s <= 'a' && nextChar >= 'A' && nextChar <= 'Z') || (s >= 'A' && s <= 'Z' && nextChar >= 'a' && nextChar <= 'z') {
				nextCharCaseChanges = true
			}
		}

		// Add an underscore before the previous character only if it's not the
		// first character, the next character case changes and we don't already
		// have an underscore there
		if i > 0 && nextCharCaseChanges && camelCase[i-1] != '_' {
			result = append(result, '_')
			result = append(result, []rune(tl)...)
		} else {
			result = append(result, []rune(tl)...)
		}
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
