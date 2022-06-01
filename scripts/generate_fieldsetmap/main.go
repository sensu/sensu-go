package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

//
// gen_gqltype
//
//   Given a Go package generates GraphQL SDL definitions for the types found
//   in the package. Takes /most/ of the busy work out of writing GraphQL
//   types!
//

var (
	pkgPath    = flag.String("pkg-path", "", "path to target package")
	apiVersion = flag.String("api-version", "", "api group & version")
	prefixes   = ArrayFlag("prefix", []string{}, "use to override the generated prefix for a fieldset func")
	output     = flag.String("o", "", "path to output file")
)

var nameExpr = regexp.MustCompile(`(\w+)Fields$`)
var listExpr = regexp.MustCompile(`^func\(\w+ \w*Resource\)`)

type fieldsetDesc struct {
	Name   string
	Kind   string
	Prefix string
}

func main() {
	flag.Parse()
	if *pkgPath == "" {
		log.Fatal("-pkg-path must be set")
	}
	if *apiVersion == "" {
		log.Fatal("-api-version must be set")
	}
	if *output == "" {
		log.Fatal("-o must be set")
	}
	prefixMap := map[string]string{}
	for _, prefix := range *prefixes {
		typename, alias, ok := strings.Cut(prefix, ":")
		if !ok {
			log.Fatal("bad prefix: " + prefix)
		}
		prefixMap[typename] = alias
	}

	fndecls, err := discoverFieldFuncs(*pkgPath)
	if err != nil {
		log.Fatal(err)
	}
	var fns []fieldsetDesc
	for _, decl := range fndecls {
		fnmatch := nameExpr.FindStringSubmatch(decl.Name.Name)
		fns = append(fns, fieldsetDesc{
			Name:   fnmatch[0],
			Kind:   fnmatch[1],
			Prefix: findPrefix(fnmatch[1], prefixMap),
		})
	}
	sort.Slice(fns, func(i, j int) bool {
		return fns[i].Name < fns[j].Name
	})

	pkgName, err := discoverPackageName(*pkgPath)
	if err != nil {
		log.Fatal(err)
	}

	outfile, err := openFile(*output)
	if err != nil {
		log.Fatal(err)
	}

	err = executeTemplate(outfile, tmplData{
		PackageName: pkgName,
		ApiVersion:  *apiVersion,
		Fieldsets:   fns,
	})
	if err != nil {
		log.Fatal(err)
	}
}

func discoverFieldFuncs(path string) (fs []*ast.FuncDecl, err error) {
	pkg, err := parsePkg(path)
	if err != nil {
		return nil, err
	}

	for _, file := range pkg.Files {
		for _, d := range file.Decls {
			f, ok := matchFieldsFunc(d)
			if !ok {
				continue
			}
			fs = append(fs, f)
		}
	}
	return
}

func discoverPackageName(path string) (string, error) {
	pkg, err := parsePkg(path)
	if err != nil {
		return "", err
	}
	return pkg.Name, nil
}

func findPrefix(typename string, mapping map[string]string) string {
	if val, ok := mapping[typename]; ok {
		return val
	}
	return snakeCase(typename)
}

func parsePkg(path string) (pkg *ast.Package, err error) {
	set := token.NewFileSet()
	pkgs, err := parser.ParseDir(set, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	for key := range pkgs {
		return pkgs[key], nil
	}
	return nil, fmt.Errorf("no package found at path: %s", path)
}

func matchFieldsFunc(d ast.Decl) (f *ast.FuncDecl, ok bool) {
	f, ok = d.(*ast.FuncDecl)
	if !ok {
		return
	}
	ok = false
	if !nameExpr.MatchString(f.Name.Name) {
		return
	}
	if strings.HasPrefix(f.Name.Name, "Test") {
		return
	}
	if !strings.HasSuffix(nodeToString(f.Type), "map[string]string") {
		return
	}
	if !listExpr.MatchString(nodeToString(f.Type)) {
		return
	}
	return f, true
}

func nodeToString(node interface{}) string {
	var buf bytes.Buffer
	printer.Fprint(&buf, token.NewFileSet(), node)
	return buf.String()
}

func openFile(path string) (*os.File, error) {
	if stat, err := os.Stat(path); err != nil && stat != nil {
		if err := os.Remove(path); err != nil {
			return nil, err
		}
	}
	f, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
	}
	return f, err
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
