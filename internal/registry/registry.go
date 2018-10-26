package registry

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sensu/sensu-go/internal/apis/meta"
	"github.com/sensu/sensu-go/internal/astutil"
)

const templateText = `package registry

// automatically generated file, do not edit!

import (
  "fmt"
  "reflect"

  "github.com/sensu/sensu-go/internal/apis/meta"
)

type registry map[meta.TypeMeta]interface{}

var typeRegistry = registry{ {{ range $index, $t := . }}
  meta.TypeMeta{APIVersion: "{{ $t.APIVersion }}", Kind: "{{ $t.Kind }}"}: {{ $t.APIVersion }}.{{ $t.Kind }}{},
  meta.TypeMeta{APIVersion: "{{ $t.APIVersion }}", Kind: "{{ lower $t.Kind }}"}: {{ $t.APIVersion }}.{{ $t.Kind }}{}, {{ end }}
}

// Resolve returns a zero-valued meta.GroupVersionKind, given a meta.TypeMeta.
// If the type does not exist, then an error will be returned.
func Resolve(mt meta.TypeMeta) (interface{}, error) {
	t, ok := typeRegistry[mt]
  if !ok {
    return nil, fmt.Errorf("type could not be found: %v", mt)
  }
  return t, nil
}
`

var (
	registryTmpl = template.Must(
		template.New("registry").
			Funcs(map[string]interface{}{
				"lower": strings.ToLower,
			}).
			Parse(templateText))
)

type templateData []meta.TypeMeta

func RegisterTypes(packagePath, outPath string) error {
	kinds, err := getPackageKinds(packagePath)
	if err != nil {
		return fmt.Errorf("couldn't get package types: %s", err)
	}

	w, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("couldn't open output for writing: %s", err)
	}

	if err := registryTmpl.Execute(w, kinds); err != nil {
		return fmt.Errorf("couldn't write registry: %s", err)
	}
	return nil
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
	sort.Slice(td, func(i, j int) bool {
		if td[i].APIVersion == td[j].APIVersion {
			return td[i].Kind < td[j].Kind
		}
		return td[i].APIVersion < td[j].APIVersion
	})
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
	if err != nil {
		return err
	}
	if fi == nil || !fi.IsDir() {
		return nil
	}
	if strings.HasPrefix(fi.Name(), ".") {
		return filepath.SkipDir
	}
	// special case for vendor directory and types package
	// we can eventually remove the special case for the types package
	if fi.Name() == "vendor" || fi.Name() == "types" {
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
