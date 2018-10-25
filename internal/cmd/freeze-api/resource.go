package main

import (
	"fmt"
	"go/doc"
	"os"
	"path"
	"path/filepath"
	"text/template"

	"github.com/sensu/sensu-go/internal/astutil"
)

const resourceTag = "+freeze-api:resource-name"

var resourceTmpl = template.Must(template.New("resource.go").Parse(`package {{ .PackageName }}

// Automatically generated file, do not edit!

/*
This file contains methods on the types in the {{ .PackageName }} package for
determining resource names.

Resource names are specified with the '+freeze-api:resource-name' special
comment, on types containing meta.TypeMeta. Resource names are specified
statically, and do not change at runtime.
*/
{{ range $i, $k := .Kinds }}
// ResourceName returns the resource name for a {{ $k.Kind }}.
// The resource name for {{ $k.Kind }} is "{{ $k.Resource }}".
func (r {{ $k.Kind }}) ResourceName() string {
	return "{{ $k.Resource }}"
}
{{ end }}
`))

type kindResource struct {
	Resource string
	Kind     string
}

func createResourceNameMethods(from, to string) error {
	pkg, err := astutil.GetPackage(from)
	if err != nil {
		return fmt.Errorf("couldn't create resource methods: %s", err)
	}

	// This is needed to extract comments from the AST and associate them with
	// types.
	docPkg := doc.New(pkg, from, 0)

	kindNames := astutil.GetKindNames(pkg)
	kr := make([]kindResource, 0, len(kindNames))
	for _, kindName := range kindNames {
		resName := resourceName(kindName, docPkg)
		kr = append(kr, kindResource{Kind: kindName, Resource: resName})
	}
	td := struct {
		PackageName string
		Kinds       []kindResource
	}{
		PackageName: path.Base(to),
		Kinds:       kr,
	}

	outPath := filepath.Join(astutil.PackagePath(to), "resource.go")

	w, err := os.OpenFile(outPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("couldn't create resource name methods: %s", err)
	}

	if err := resourceTmpl.Execute(w, td); err != nil {
		return fmt.Errorf("couldn't create resource name methods: %s", err)
	}

	return nil
}

func resourceName(kindName string, pkg *doc.Package) string {
	return astutil.ParseCommentTags(kindName, pkg)[resourceTag]
}
