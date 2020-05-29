package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"text/template"
	"unicode"
	"unicode/utf8"
)

var (
	tmplPath = flag.String("t", "", "Path to template file")
	output   = flag.String("o", "", "Path to output file")
	typeRe   = regexp.MustCompile(`^type ([A-Z].+) struct \{`)
)

type tmplData struct {
	TypeNames []string
	Comment   string
}

// kebabCase is like snakeCase, but underscores are converted to dashes
func kebabCase(camelCase string) string {
	return strings.ReplaceAll(snakeCase(camelCase), "_", "-")
}

func snakeCase(camelCase string) string {
	result := make([]rune, 0)
	for i, s := range camelCase {
		tl := strings.ToLower(string(s))

		// Treat acronyms as single-word, e.g. LDAP -> ldap
		var nextCharCaseChanges bool
		if i+1 < len(camelCase) {
			nextChar, _ := utf8.DecodeRune([]byte{camelCase[i+1]})
			// Check if the next character case differs from the current character
			if (unicode.IsLower(s) && unicode.IsUpper(nextChar)) || (unicode.IsUpper(s) && unicode.IsLower(nextChar)) {
				nextCharCaseChanges = true
			}
		}

		// Add an underscore before the previous character only if it's not the
		// first character, the next character case changes and we don't already
		// have an underscore in the result rune
		if i > 0 && nextCharCaseChanges && result[len(result)-1] != '_' {
			// Prepend the underscore if the next character is lowercase, otherwise
			// append it
			if unicode.IsUpper(s) {
				result = append(result, '_')
				result = append(result, []rune(tl)...)
			} else if unicode.IsLower(s) {
				result = append(result, []rune(tl)...)
				result = append(result, '_')
			}
		} else {
			result = append(result, []rune(tl)...)
		}
	}
	return string(result)
}

func receiver(name string) string {
	if name == "" {
		panic("can't resolve receiver name")
	}
	letter := strings.ToLower(name[:1])
	return fmt.Sprintf("%s *%s", letter, name)
}

func rvar(name string) string {
	if name == "" {
		panic("can't resolve receiver name")
	}
	return strings.ToLower(name[:1])
}

func storeSuffix(name string) string {
	return fmt.Sprintf("%ss", snakeCase(name))
}

func rbacName(name string) string {
	return fmt.Sprintf("%ss", snakeCase(name))
}

func main() {
	flag.Parse()
	tmpl, err := template.New(*tmplPath).Funcs(template.FuncMap{
		"snakeCase":   snakeCase,
		"receiver":    receiver,
		"rvar":        rvar,
		"storeSuffix": storeSuffix,
		"rbacName":    rbacName,
		"kebabCase":   kebabCase,
	}).ParseFiles(*tmplPath)

	if err != nil {
		log.Fatalf("fatal error parsing %s: %s", *tmplPath, err)
	}
	typeNames, err := discoverTypeNames()
	if err != nil {
		log.Fatalf("fatal error discovering types: %s", err)
	}
	tmplData := tmplData{
		TypeNames: typeNames,
		Comment:   "automatically generated file, do not edit!",
	}
	out, err := os.Create(*output)
	if err != nil {
		log.Fatalf("fatal error creating typemap.go: %s", err)
	}
	if err := tmpl.Execute(out, tmplData); err != nil {
		log.Fatalf("fatal error generating typemap.go: %s", err)
	}
}

func discoverTypeNames() ([]string, error) {
	var typeNames []string
	doc, err := exec.Command("go", "doc", "-all", ".").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s: %s", string(doc), err)
	}
	scanner := bufio.NewScanner(bytes.NewReader(doc))
	for scanner.Scan() {
		line := scanner.Bytes()
		matches := typeRe.FindSubmatch(line)
		if len(matches) > 1 {
			// capturing group match in matches[1]
			typeNames = append(typeNames, string(matches[1]))
		}
	}
	return typeNames, scanner.Err()
}
