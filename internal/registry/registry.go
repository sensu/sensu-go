package registry

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/sensu/sensu-go/internal/api"
	metav1 "github.com/sensu/sensu-go/internal/apis/meta/v1"
	"github.com/sensu/sensu-go/internal/astutil"
)

const templateText = `package registry

// automatically generated file, do not edit!

import (
  "fmt"
  "reflect"

  metav1 "github.com/sensu/sensu-go/internal/apis/meta/v1"
  {{ range $i, $import := imports . }}{{ $import }}
  {{ end }}
)

type registry map[metav1.TypeMeta]interface{}

var typeRegistry = registry{ {{ range $index, $t := . }}
  metav1.TypeMeta{APIVersion: "{{ $t.APIVersion }}", Kind: "{{ $t.Kind }}"}: {{ replace $t.APIVersion "/" "" 1 }}.{{ $t.Kind }}{},
  metav1.TypeMeta{APIVersion: "{{ $t.APIVersion }}", Kind: "{{ lower $t.Kind }}"}: {{ replace $t.APIVersion "/" "" 1 }}.{{ $t.Kind }}{}, {{ end }}
}

// Resolve returns a zero-valued metav1.GroupVersionKind, given a metav1.TypeMeta.
// If the type does not exist, then an error will be returned.
func Resolve(mt metav1.TypeMeta) (interface{}, error) {
	t, ok := typeRegistry[mt]
  if !ok {
    return nil, fmt.Errorf("type could not be found: %v", mt)
  }
  return t, nil
}
`

func imports(kinds []metav1.TypeMeta) (result []string) {
	set := map[string]struct{}{}
	for _, kind := range kinds {
		var imp string
		if strings.Contains(kind.APIVersion, "/") {
			imp = fmt.Sprintf(`%s "github.com/sensu/sensu-go/apis/%s"`, strings.Replace(kind.APIVersion, "/", "", -1), kind.APIVersion)
		} else {
			imp = fmt.Sprintf(`"github.com/sensu/sensu-go/internal/apis/%s"`, kind.APIVersion)
		}
		fmt.Printf("%q\n", imp)
		set[imp] = struct{}{}
	}
	for k := range set {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}

var (
	registryTmpl = template.Must(
		template.New("registry").
			Funcs(map[string]interface{}{
				"lower":    strings.ToLower,
				"replace":  strings.Replace,
				"contains": strings.Contains,
				"imports":  imports,
			}).
			Parse(templateText))
)

type templateData []metav1.TypeMeta

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
// sensu-go kinds (structs that embed metav1.TypeMeta).
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
			td = append(td, metav1.TypeMeta{APIVersion: pkg.Name, Kind: kind})
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
	if fi.Name() == "meta" {
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
	for _, v := range packages {
		key := apiName(path)
		w.packages[key] = v
	}
	return nil
}

func apiName(pth string) string {
	parts := filepath.SplitList(pth)
	if version, err := api.ParseVersion(parts[len(parts)-1]); err != nil {
		return filepath.Base(pth)
	} else {
		return fmt.Sprintf("%s/%s", parts[len(parts)-2], version.String())
	}
}
