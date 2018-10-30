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

	metav1 "github.com/sensu/sensu-go/apis/meta/v1"
	"github.com/sensu/sensu-go/internal/api"
	"github.com/sensu/sensu-go/internal/astutil"
)

const fileFlags = os.O_CREATE | os.O_TRUNC | os.O_WRONLY

const templateText = `package registry

// automatically generated file, do not edit!

import (
	"fmt"
	"reflect"
	
	metav1 "github.com/sensu/sensu-go/apis/meta/v1"
	{{ range $i, $import := imports . }}{{ $import }}
	{{ end }}
)

type registry map[metav1.TypeMeta]interface{}

var typeRegistry = registry{ {{ range $index, $t := . }}
	metav1.TypeMeta{APIVersion: "{{ $t.APIVersion }}", Kind: "{{ $t.Kind }}"}: {{ replace $t.APIVersion "/" "" 1 }}.{{ $t.Kind }}{},
	metav1.TypeMeta{APIVersion: "{{ $t.APIVersion }}", Kind: "{{ lower $t.Kind }}"}: {{ replace $t.APIVersion "/" "" 1 }}.{{ $t.Kind }}{}, {{ end }}
}

func init() {
	for k, v := range typeRegistry {
		r, ok := v.(interface{ResourceName() string})
		if ok {
			newKey := metav1.TypeMeta{
				APIVersion: k.APIVersion,
				Kind: r.ResourceName(),
			}
			typeRegistry[newKey] = v
		}
	}
}

// Resolve returns a zero-valued sensu object, given a metav1.TypeMeta.
// If the type does not exist, then an error will be returned.
func Resolve(mt metav1.TypeMeta) (interface{}, error) {
	t, ok := typeRegistry[mt]
	if !ok {
	  return nil, fmt.Errorf("type could not be found: %v", mt)
	}
	return t, nil
}

// ResolveSlice returns a zero-valued slice of sensu objects, given a
// meta.TypeMeta. If the type does not exist, then an error will be returned.
func ResolveSlice(mt metav1.TypeMeta) (interface{}, error) {
	t, err := Resolve(mt)
	if err != nil {
		return nil, err
	}
	return reflect.Indirect(reflect.New(reflect.SliceOf(reflect.TypeOf(t)))).Interface(), nil
}
`

const testText = `package registry

import (
	"fmt"
	"reflect"
	"testing"

	metav1 "github.com/sensu/sensu-go/apis/meta/v1"
	"github.com/sensu/sensu-go/internal/api"
)

func TestRegistryResourceAliases(t *testing.T) {
	for key, kind := range typeRegistry {
		if api.IsInternal(key.APIVersion) {
			continue
		}
		r := kind.(interface{ ResourceName() string })
		t.Run(fmt.Sprintf("%s -> %s", key.Kind, r.ResourceName()), func(t *testing.T) {
			_, ok := typeRegistry[metav1.TypeMeta{APIVersion: key.APIVersion, Kind: r.ResourceName()}]
			if !ok {
				t.Fatalf("%v resource missing", key)
			}
		})
	}
}

func TestResolveSlice(t *testing.T) {
	for key, kind := range typeRegistry {
		t.Run(fmt.Sprintf("slice of %s", key.Kind), func(t *testing.T) {
			defer func() {
				if e := recover(); e != nil {
					t.Fatal(e)
				}
			}()
			slice, err := ResolveSlice(key)
			if err != nil {
				t.Fatal(err)
			}
			// Will panic if ResolveSlice is broken
			reflect.Append(reflect.ValueOf(slice), reflect.ValueOf(kind))
		})
	}
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

	w, err := os.OpenFile(filepath.Join(outPath, "registry.go"), fileFlags, 0644)
	if err != nil {
		return fmt.Errorf("couldn't open output for writing: %s", err)
	}

	if err := registryTmpl.Execute(w, kinds); err != nil {
		return fmt.Errorf("couldn't write registry: %s", err)
	}

	x, err := os.OpenFile(filepath.Join(outPath, "registry_test.go"), fileFlags, 0644)
	if err != nil {
		return fmt.Errorf("couldn't open output for writing: %s", err)
	}

	_, err = fmt.Fprint(x, testText)
	return err
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
	for apiGroup, pkg := range packages {
		if strings.HasSuffix(pkg.Name, "_test") {
			continue
		}
		kinds := astutil.GetKindNames(pkg)
		for _, kind := range kinds {
			td = append(td, metav1.TypeMeta{APIVersion: apiGroup, Kind: kind})
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

func apiName(pth string) (result string) {
	parts := strings.Split(filepath.ToSlash(pth), "/")
	if version, err := api.ParseVersion(parts[len(parts)-1]); err != nil {
		return filepath.Base(pth)
	} else {
		return fmt.Sprintf("%s/%s", parts[len(parts)-2], version.String())
	}
}
