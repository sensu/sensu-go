package astutil

import (
	"errors"
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
)

// PackagePath returns the filesystem path of supplied the Go package.
func PackagePath(path string) string {
	return filepath.Join(build.Default.GOPATH, "src", path)
}

func removeTestPackages(packages map[string]*ast.Package) {
	for k := range packages {
		if strings.HasSuffix(k, "_test") {
			delete(packages, k)
		}
	}
}

// GetPacakge parses the given Go package, and returns an *ast.Package, along
// with any error encountered.
func GetPackage(pkg string) (*ast.Package, error) {
	path := PackagePath(pkg)
	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("error parsing package: %s", err)
	}
	removeTestPackages(packages)
	if len(packages) > 1 {
		return nil, errors.New("too many 'from' packages")
	}
	for _, v := range packages {
		return v, nil
	}
	return nil, errors.New("no packages found")
}
