package main

import (
	"fmt"
	"go/ast"
	"os"
	"path"
	"reflect"
	"text/template"

	"github.com/sensu/sensu-go/internal/astutil"
)

const conversionPackage = "github.com/sensu/sensu-go/internal/conversion"

type templateData struct {
	// VersionedPackage is the name of the versioned package.
	VersionedPackage string

	// InternalPackage is the name of the internal package.
	InternalPackage string

	// InternalPackagePath is the fully qualified package path of InternalPackage
	InternalPackagePath string

	// VersionedPackagePath is the fully qualified package path of VersionedPackage
	VersionedPackagePath string

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

const conversionInitFuncs = `package conversion

import (
	"{{ .InternalPackagePath }}"
)
{{ $vPkg := .VersionedPackage }}
{{ $iPkg := .InternalPackage }}
{{ range $i, $t := .Types }}
{{ $internalToVersioned := conversionFuncName $iPkg $vPkg $t.TypeName }}
{{ $versionedToInternal := conversionFuncName $vPkg $iPkg $t.TypeName }}
func init() {
	registry[key{
		SourceAPIVersion: "{{ $iPkg }}",
		DestAPIVersion: "{{ $iPkg }}/{{ $vPkg}}",
		Kind: "{{ $t.TypeName }}",
	}] = {{ $iPkg }}.{{ $internalToVersioned }}

	registry[key{
		SourceAPIVersion: "{{ $iPkg }}/{{ $vPkg}}",
		DestAPIVersion: "{{ $iPkg }}",
		Kind: "{{ $t.TypeName }}",
	}] = {{ $iPkg }}.{{ $versionedToInternal }}
}
{{ end }}
`

const converterTmplStr = `package {{ .InternalPackage }}

import (
	"unsafe"

	"{{ .VersionedPackagePath }}"
)
{{ $vPkg := .VersionedPackage }}
{{ $iPkg := .InternalPackage }}
{{ range $i, $t := .Types }}
{{ $internalToVersioned := conversionFuncName $iPkg $vPkg $t.TypeName }}
{{ $versionedToInternal := conversionFuncName $vPkg $iPkg $t.TypeName }}
{{ $internalType := $t.TypeName }}
{{ $versionedType := typeName $vPkg $t.TypeName }}
func {{ $internalToVersioned }}(dst, src interface{}) error {
	dstp := dst.(*{{ $versionedType }})
	srcp := src.(*{{ $internalType }})
	{{ if $t.Simple }}
	*dstp = *(*{{ $versionedType }})(unsafe.Pointer(srcp))
	{{ else }}
	panic("complex conversions not supported yet")
	{{ end }}
	return nil
}

func {{ $versionedToInternal }}(dst, src interface{}) error {
	dstp := dst.(*{{ $internalType }})
	srcp := src.(*{{ $versionedType }})
	{{ if $t.Simple }}
	*dstp = *(*{{ $internalType }})(unsafe.Pointer(srcp))
	{{ else }}
	panic("complex conversions not supported yet")
	{{ end }}
	return nil
}
{{ end }}`

const converterTestTmplStr = `package {{ .InternalPackage }}

import (
	"reflect"
	"testing"

	fuzz "github.com/google/gofuzz"
	"{{ .VersionedPackagePath }}"
)
{{ $vPkg := .VersionedPackage }}
{{ $iPkg := .InternalPackage }}
{{ range $i, $t := .Types }}
{{ $internalToVersioned := conversionFuncName $iPkg $vPkg $t.TypeName }}
{{ $versionedToInternal := conversionFuncName $vPkg $iPkg $t.TypeName }}
{{ $internalType := $t.TypeName }}
{{ $versionedType := typeName $vPkg $t.TypeName }}
func Test_Convert_{{ $internalToVersioned }}_And_{{ $versionedToInternal }}(t *testing.T) {
	var v1, v2 {{ $internalType }}
	var v3 {{ $versionedType }}
	fuzzer := fuzz.New().NilChance(0)
	fuzzer.Fuzz(&v1)
	if err := {{ $internalToVersioned }}(&v3, &v1); err != nil {
		t.Fatal(err)
	}
	if err := {{ $versionedToInternal }}(&v2, &v3); err != nil {
		t.Fatal(err)
	}
	{{ if $t.Simple }}
	if !reflect.DeepEqual(v1, v2) {
		t.Fatal("values not equal")
	}{{ end }}
}{{ end }}
`

var templateFuncs = map[string]interface{}{
	"conversionFuncName": func(fromPkg, versionedPkg, typename string) string {
		return fmt.Sprintf("Convert_%s_%s_To_%s_%s", fromPkg, typename, versionedPkg, typename)
	},
	"typeName": func(pkg, typename string) string {
		return fmt.Sprintf("%s.%s", pkg, typename)
	},
}

var (
	converterTmpl           = template.Must(template.New("converter").Funcs(templateFuncs).Parse(converterTmplStr))
	converterTestsTmpl      = template.Must(template.New("converter_test").Funcs(templateFuncs).Parse(converterTestTmplStr))
	conversionInitFuncsTmpl = template.Must(template.New("pkginits").Funcs(templateFuncs).Parse(conversionInitFuncs))
)

func createConverters(internal, versioned string) error {
	internalPackage, err := astutil.GetPackage(internal)
	if err != nil {
		return err
	}

	versionedPackage, err := astutil.GetPackage(versioned)
	if err != nil {
		return err
	}

	internalTypes := astutil.GetKinds(internalPackage)
	versionedTypes := astutil.GetKinds(versionedPackage)

	td := templateData{
		VersionedPackage:     path.Base(versioned),
		InternalPackage:      path.Base(internal),
		VersionedPackagePath: versioned,
		InternalPackagePath:  internal,
	}

	for _, typeName := range astutil.GetKindNames(internalPackage) {
		internalType := internalTypes[typeName]
		versionedType, ok := versionedTypes[typeName]
		if !ok {
			continue
		}
		simple := typesEquivalent(internalType, versionedType)
		td.Types = append(td.Types, conversionType{
			Simple:   simple,
			TypeName: typeName,
		})
	}

	outPaths := []string{
		path.Join(
			astutil.PackagePath(internal),
			fmt.Sprintf("converters_%s_generated.go", path.Base(versioned))),
		path.Join(
			astutil.PackagePath(internal),
			fmt.Sprintf("converters_%s_generated_test.go", path.Base(versioned))),
		path.Join(
			astutil.PackagePath(conversionPackage),
			fmt.Sprintf("registry_%s_generated.go", path.Base(versioned))),
	}

	tmpls := []*template.Template{
		converterTmpl,
		converterTestsTmpl,
		conversionInitFuncsTmpl,
	}

	for i := range outPaths {
		if err := writeTemplate(outPaths[i], tmpls[i], td); err != nil {
			return err
		}
	}

	return nil
}

func writeTemplate(path string, tmpl *template.Template, data interface{}) (err error) {
	w, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("couldn't create %s: %s", path, err)
	}
	defer func() {
		if err == nil {
			err = w.Close()
		}
	}()
	return tmpl.Execute(w, data)
}

func typesEquivalent(a, b *ast.TypeSpec) bool {
	t1 := reflect.Indirect(reflect.ValueOf(a.Type)).Type()
	t2 := reflect.Indirect(reflect.ValueOf(b.Type)).Type()

	return t1 == t2
}
