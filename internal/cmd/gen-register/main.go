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
)

type templateData []meta.TypeMeta

var (
	packagePath = flag.String("pkg", "", "Path to package to generate registry for")
	tmplPath    = flag.String("t", "", "Path to template file")
	outPath     = flag.String("o", "", "Output path")
)

func main() {
	log.SetFlags(log.Lshortfile)
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

	types, err := getPackageTypes(*packagePath)
	if err != nil {
		log.Fatalf("couldn't get package types: %s", err)
	}

	w, err := os.OpenFile(*outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("couldn't open output for writing: %s", err)
	}

	if err := tmpl.Execute(w, types); err != nil {
		log.Fatalf("couldn't write registry: %s", err)
	}
}

func getPackageTypes(path string) (templateData, error) {
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

func scanPackages(packages map[string]*ast.Package) templateData {
	td := make(templateData, 0)
	for _, pkg := range packages {
		if strings.HasSuffix(pkg.Name, "_test") {
			continue
		}
		data := scanPackage(pkg)
		td = append(td, data...)
	}
	return td
}

func scanPackage(pkg *ast.Package) templateData {
	result := make(templateData, 0)
	kinds := getTypeKinds(pkg)
	for _, kind := range kinds {
		result = append(result, meta.TypeMeta{APIVersion: pkg.Name, Kind: kind})
	}
	return result
}

type walker struct {
	packages map[string]*ast.Package
	fset     *token.FileSet
}

func getTypeKinds(pkg *ast.Package) []string {
	result := make([]string, 0)
	for _, f := range pkg.Files {
		for _, decl := range f.Decls {
			gendecl, ok := decl.(*ast.GenDecl)
			if !ok || gendecl.Tok != token.TYPE {
				continue
			}
			for _, spec := range gendecl.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || !ts.Name.IsExported() {
					continue
				}
				strukt, ok := ts.Type.(*ast.StructType)
				if !ok {
					continue
				}
				for _, field := range strukt.Fields.List {
					if len(field.Names) != 0 {
						// not embedded
						continue
					}
					expr, ok := field.Type.(*ast.SelectorExpr)
					if !ok {
						continue
					}
					if expr.Sel.Name == "TypeMeta" {
						result = append(result, ts.Name.Name)
					}
				}
			}
		}
	}
	return result
}

func (w *walker) walk(path string, fi os.FileInfo, err error) error {
	if !fi.IsDir() {
		return nil
	}
	if fi.IsDir() && fi.Name() == "internal" {
		return filepath.SkipDir
	}
	if fi.IsDir() && strings.HasPrefix(fi.Name(), ".") {
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
