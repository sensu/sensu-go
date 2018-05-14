package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/doc"
	"go/parser"
	"go/printer"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
)

// parseSourceFlag parses the "-source" flag value. It must have "import/path".VariableName format.
func parseSourceFlag(sourceFlag string) (importPath, variableName string, err error) {
	// Parse sourceFlag as a Go expression, albeit a strange one:
	//
	// 	"import/path".VariableName
	//
	e, err := parser.ParseExpr(sourceFlag)
	if err != nil {
		return "", "", fmt.Errorf("invalid format, failed to parse %q as a Go expression", sourceFlag)
	}
	se, ok := e.(*ast.SelectorExpr)
	if !ok {
		return "", "", fmt.Errorf("invalid format, expression %v is not a selector expression but %T", sourceFlag, e)
	}
	importPath, err = stringValue(se.X)
	if err != nil {
		return "", "", fmt.Errorf("invalid format, expression %v is not a properly quoted Go string: %v", stringifyAST(se.X), err)
	}
	variableName = se.Sel.Name
	return importPath, variableName, nil
}

// stringValue returns the string value of string literal e.
func stringValue(e ast.Expr) (string, error) {
	lit, ok := e.(*ast.BasicLit)
	if !ok {
		return "", fmt.Errorf("not a string, but %T", e)
	}
	if lit.Kind != token.STRING {
		return "", fmt.Errorf("not a string, but %v", lit.Kind)
	}
	return strconv.Unquote(lit.Value)
}

// parseTagFlag parses the "-tag" flag value. It must be a single build tag.
func parseTagFlag(tagFlag string) (tag string, err error) {
	tags := strings.Fields(tagFlag)
	if len(tags) != 1 {
		return "", fmt.Errorf("%q is not a valid single build tag, but %q", tagFlag, tags)
	}
	return tags[0], nil
}

// lookupNameAndComment imports package using provided build context, and
// returns the package name and variable comment.
func lookupNameAndComment(bctx build.Context, importPath, variableName string) (packageName, variableComment string, err error) {
	bpkg, err := bctx.Import(importPath, "", 0)
	if err != nil {
		return "", "", fmt.Errorf("can't import package %q: %v", importPath, err)
	}
	dpkg, err := docPackage(bpkg)
	if err != nil {
		return "", "", fmt.Errorf("can't get godoc of package %q: %v", importPath, err)
	}
	for _, v := range dpkg.Vars {
		if len(v.Names) == 1 && v.Names[0] == variableName {
			variableComment = strings.TrimSuffix(v.Doc, "\n")
			break
		}
	}
	return bpkg.Name, variableComment, nil
}

func stringifyAST(node interface{}) string {
	var buf bytes.Buffer
	err := printer.Fprint(&buf, token.NewFileSet(), node)
	if err != nil {
		return "printer.Fprint error: " + err.Error()
	}
	return buf.String()
}

// TODO: Keep in sync or unify with github.com/shurcooL/cmd/gorepogen/main.go.
// TODO: See if these can be cleaned up.

func docPackage(bpkg *build.Package) (*doc.Package, error) {
	apkg, err := astPackage(bpkg)
	if err != nil {
		return nil, err
	}
	return doc.New(apkg, bpkg.ImportPath, 0), nil
}

func astPackage(bpkg *build.Package) (*ast.Package, error) {
	// TODO: Either find a way to use golang.org/x/tools/importer (from Go 1.4~ or older, it no longer exists as of Go 1.6) directly, or do file AST parsing in parallel like it does.
	filenames := append(bpkg.GoFiles, bpkg.CgoFiles...)
	files := make(map[string]*ast.File, len(filenames))
	fset := token.NewFileSet()
	for _, filename := range filenames {
		fileAst, err := parser.ParseFile(fset, filepath.Join(bpkg.Dir, filename), nil, parser.ParseComments)
		if err != nil {
			return nil, err
		}
		files[filename] = fileAst // TODO: Figure out if filename or full path are to be used (the key of this map doesn't seem to be used anywhere).
	}
	return &ast.Package{Name: bpkg.Name, Files: files}, nil
}
