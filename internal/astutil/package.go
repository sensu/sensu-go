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

func packagePath(path string) string {
	return filepath.Join(build.Default.GOPATH, "src", path)
}

func removeTestPackages(packages map[string]*ast.Package) {
	for k := range packages {
		if strings.HasSuffix(k, "_test") {
			delete(packages, k)
		}
	}
}

func GetPackage(pkg string) (*ast.Package, error) {
	path := packagePath(pkg)
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
