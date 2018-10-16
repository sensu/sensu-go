package main

import (
	"flag"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/sensu/sensu-go/internal/apis/meta"
	"github.com/sensu/sensu-go/internal/astutil"
)

type templateData []meta.TypeMeta

var (
	packagePath = flag.String("pkg", "", "Path to package to generate registry for")
	tmplPath    = flag.String("t", "", "Path to template file")
	outPath     = flag.String("o", "", "Output path")
)

func main() {
	log.SetFlags(log.Lshortfile)
	flag.Usage = func() {
		w := flag.CommandLine.Output()
		fmt.Fprintf(w, "%s: Generate a type registry for sensu-go API types.\n", os.Args[0])
		fmt.Fprintln(w)
		fmt.Fprintln(w, "Example usage: gen-register -pkg github.com/sensu/sensu-go -t register.go.tmpl -o register.go")
		flag.PrintDefaults()
	}
	flag.Parse()
	if *packagePath == "" {
		log.Fatal("no package path supplied (-pkg)")
	}
	if *tmplPath == "" {
		log.Fatal("no template path supplied (-t)")
	}
	if *outPath == "" {
		log.Fatal("no output path supplied (-o)")
	}

	tmpl, err := template.ParseFiles(*tmplPath)
	if err != nil {
		log.Fatalf("couldn't read template: %s", err)
	}

	kinds, err := getPackageKinds(*packagePath)
	if err != nil {
		log.Fatalf("couldn't get package types: %s", err)
	}

	w, err := os.OpenFile(*outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("couldn't open output for writing: %s", err)
	}

	if err := tmpl.Execute(w, kinds); err != nil {
		log.Fatalf("couldn't write registry: %s", err)
	}
}

// getPackageKinds recursively traverses the package and all sub-packages for
// sensu-go kinds (structs that embed meta.TypeMeta).
func getPackageKinds(path string) (templateData, error) {
	root := filepath.Join(build.Default.GOPATH, "src", path)
	walker := &walker{
		fset:     token.NewFileSet(),
		packages: make(map[string]*ast.Package),
	}
	if err := filepath.Walk(root, walker.walk); err != nil {
		return nil, fmt.Errorf("couldn't walk filesystem: %s", err)
	}
	td := scanPackages(walker.packages)
	return td, nil
}

// scanPackages scans all of the collected packages for kinds and collects them
// into a slice.
func scanPackages(packages map[string]*ast.Package) templateData {
	td := make(templateData, 0)
	for _, pkg := range packages {
		if strings.HasSuffix(pkg.Name, "_test") {
			continue
		}
		kinds := astutil.GetKindNames(pkg)
		for _, kind := range kinds {
			td = append(td, meta.TypeMeta{APIVersion: pkg.Name, Kind: kind})
		}
	}
	return td
}

// walker walks a filesystem recursively, looking for Go packages to parse.
type walker struct {
	packages map[string]*ast.Package
	fset     *token.FileSet
}

func (w *walker) walk(path string, fi os.FileInfo, err error) error {
	if fi == nil || !fi.IsDir() {
		return nil
	}
	if strings.HasPrefix(fi.Name(), ".") {
		return filepath.SkipDir
	}
	if fi.Name() == "vendor" {
		return filepath.SkipDir
	}
	packages, err := parser.ParseDir(w.fset, path, nil, 0)
	if err != nil {
		return fmt.Errorf("couldn't parse directory: %s", err)
	}
	for k, v := range packages {
		w.packages[k] = v
	}
	return nil
}
