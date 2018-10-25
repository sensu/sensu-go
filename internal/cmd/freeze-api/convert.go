package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"text/template"

	"github.com/sensu/sensu-go/internal/astutil"
)

type templateData struct {
	// ToPackage is the name of the destination package.
	ToPackage string

	// FromPackage is the name of the source package.
	FromPackage string

	// ImportPackage is the fully qualified package path of ToPackage
	ImportPackage string

	// Types is a list of types to generate converters for
	Types []conversionType
}

type conversionType struct {
	// TypeName is the name of the type being converted.
	TypeName string

	// Simple informs the template whether or not to do a simple conversion.
	// If the two types are structurally equivalent, then the type will be
	// converted with an unsafe pointer conversion.
	Simple bool

	// TODO(echlebek): support complex conversions
	ComplexFields map[string]reflect.Kind
}

const converterTmplStr = `package {{ .FromPackage }}

import (
	"unsafe"

	"{{ .ImportPackage }}"
)
{{ $toPackage := .ToPackage }}{{ range $idx, $t := .Types }}
// ConvertTo converts a *{{ $t.TypeName }} to a *{{ $toPackage }}.{{ $t.TypeName }}.
// It panics if the to parameter is not a *{{ $toPackage }}.{{ $t.TypeName }}.
func (r *{{ $t.TypeName }}) ConvertTo(to interface{}) {
	ptr := to.(*{{ $toPackage }}.{{ $t.TypeName }})
	convert_{{ $t.TypeName }}_To_{{ $toPackage }}_{{ $t.TypeName }}(r, ptr)
}

var convert_{{ $t.TypeName }}_To_{{ $toPackage }}_{{ $t.TypeName}} = func(from *{{ $t.TypeName }}, to *{{ $toPackage}}.{{ $t.TypeName }}) {
	{{ if $t.Simple }}*to = *(*{{ $toPackage }}.{{ $t.TypeName }})(unsafe.Pointer(from)){{ else }}panic("complex conversion not supported yet"){{ end }}
}

// ConvertFrom converts the receiver to a *{{ $toPackage }}.{{ $t.TypeName }}.
// It panics if the from parameter is not a *{{ $toPackage }}.{{ $t.TypeName }}.
func (r *{{ $t.TypeName}}) ConvertFrom(from interface{}) {
	ptr := from.(*{{ $toPackage }}.{{ $t.TypeName }})
	convert_{{ $toPackage }}_{{ $t.TypeName}}_To_{{ $t.TypeName}}(ptr, r)
}

var convert_{{ $toPackage }}_{{ $t.TypeName}}_To_{{ $t.TypeName}} = func(from *{{ $toPackage }}.{{ $t.TypeName }}, to *{{ $t.TypeName }}) {
	{{ if $t.Simple }}*to = *(*{{ $t.TypeName }})(unsafe.Pointer(from)){{ else }}panic("complex conversion not supported yet"){{end}}
}
{{ end }}
`

const converterTestTmplStr = `package {{ .FromPackage }}

import (
	"reflect"
	"testing"

	fuzz "github.com/google/gofuzz"
	"{{ .ImportPackage }}"
)
{{ $toPackage := .ToPackage }}{{ range $i, $t := .Types }}
func Test_convert_{{ $t.TypeName }}_To_{{ $toPackage }}_{{ $t.TypeName}}(t *testing.T) {
	var v1, v2 {{ $t.TypeName }}
	var v3 {{ $toPackage }}.{{ $t.TypeName }}
	fuzzer := fuzz.New().NilChance(0)
	fuzzer.Fuzz(&v1)
	v1.ConvertTo(&v3)
	v2.ConvertFrom(&v3)
	{{ if $t.Simple }}if !reflect.DeepEqual(v1, v2) {
		t.Fatal("values not equal")
	}{{ end }}
}{{ end }}
`

var (
	converterTmpl      = template.Must(template.New("converter").Parse(converterTmplStr))
	converterTestsTmpl = template.Must(template.New("converter_test").Parse(converterTestTmplStr))
)

func createConverters(from, to string) error {
	fromPackage, err := astutil.GetPackage(from)
	if err != nil {
		return err
	}

	toPackage, err := astutil.GetPackage(to)
	if err != nil {
		return err
	}

	fromTypes := astutil.GetKinds(fromPackage)
	toTypes := astutil.GetKinds(toPackage)

	td := templateData{
		ToPackage:     path.Base(to),
		FromPackage:   path.Base(from),
		ImportPackage: to,
	}

	for _, typeName := range astutil.GetKindNames(fromPackage) {
		fromType := fromTypes[typeName]
		toType, ok := toTypes[typeName]
		if !ok {
			fmt.Println(typeName)
			continue
		}
		simple := typesEquivalent(fromType, toType)
		td.Types = append(td.Types, conversionType{
			Simple:   simple,
			TypeName: typeName,
		})
	}

	outPath := path.Join(packagePath(from), "converters.go")
	w, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("couldn't create converters.go: %s", err)
	}
	defer w.Close()

	testOutPath := path.Join(packagePath(from), "converters_test.go")
	x, err := os.OpenFile(testOutPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("couldn't create converters_test.go: %s", err)
	}
	defer x.Close()

	if err := converterTmpl.Execute(w, td); err != nil {
		return err
	}

	return converterTestsTmpl.Execute(x, td)
}

func typesEquivalent(a, b *ast.TypeSpec) bool {
	t1 := reflect.Indirect(reflect.ValueOf(a.Type)).Type()
	t2 := reflect.Indirect(reflect.ValueOf(b.Type)).Type()

	return t1 == t2
}

func packagePath(path string) string {
	return filepath.Join(build.Default.GOPATH, "src", path)
}
