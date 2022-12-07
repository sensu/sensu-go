package main

import (
	"flag"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"regexp"
	"sort"
	gostrings "strings"
	"unicode"

	gqlast "github.com/graphql-go/graphql/language/ast"
	gqlkind "github.com/graphql-go/graphql/language/kinds"
	gqlprinter "github.com/graphql-go/graphql/language/printer"
	"github.com/sensu/sensu-go/util/strings"
	"golang.org/x/mod/semver"
)

//
// gen_gqltype
//
//   Given a Go package generates GraphQL SDL definitions for the types found
//   in the package. Takes /most/ of the busy work out of writing GraphQL
//   types!
//

var (
	pkgPath   = flag.String("pkg-path", "", "path to target package; may be a path to the source or a go import path from a required module")
	output    = flag.String("o", "", "path to output file")
	types     = flag.String("types", "", "comma separated list of types to export; optional")
	noPkg     = flag.Bool("no-pkg", false, "by default types are prefix'd with its package name; use this flag to diable this behaviour")
	camelCase = flag.Bool("camel-case", false, "converts field names to camelCase; when false the field's json tag is used as the name")
)

func main() {
	flag.Parse()
	if *pkgPath == "" {
		log.Fatal("-pkg-path must be set")
	}
	if *output == "" {
		log.Fatal("-o must be set")
	}
	validTypes := []string{}
	if *types != "" {
		validTypes = gostrings.Split(*types, ",")
	}

	srcDir := *pkgPath
	// check if exists, otherwise attempt to downlaod go module source

	if _, err := os.Stat(*pkgPath); err != nil {
		dir, cleanup, modErr := WithGoModuleSource(*pkgPath)
		if modErr != nil {
			log.Fatalf("could not find package source %v\n", modErr)
		}
		defer cleanup()
		log.Printf("go module source cloned to %s\n", dir)
		srcDir = dir
	}
	structs, err := discoverStructs(srcDir)
	if err != nil {
		log.Fatal(err)
	}
	prefix := ""
	if !*noPkg {
		prefix = pathToPackageName(*pkgPath)
	}
	imports, err := discoverImports(srcDir)
	if err != nil {
		log.Fatal(err)
	}

	// Sort structs alphabetically so we get some semblance of predictable
	// output.
	sort.Slice(structs, func(i, j int) bool {
		return structs[i].Name.Name < structs[j].Name.Name
	})

	translatorCfg := &translatorCfg{
		imports:   imports,
		camelCase: *camelCase,
		pkgPrefix: !*noPkg,
		pkgName:   prefix,
	}
	doc := &gqlast.Document{Kind: gqlkind.Document}
	for _, s := range structs {
		name := ""
		if s.Name != nil {
			name = s.Name.Name
		}

		// skip private structs
		if len(name) == 0 || unicode.IsLower(rune(name[0])) {
			continue
		}

		// ignore type if it was not specified by the user
		if len(validTypes) > 0 && !strings.InArray(name, validTypes) {
			continue
		}

		t := s.Type.(*ast.StructType)
		def := &gqlast.ObjectDefinition{
			Kind: gqlkind.ObjectDefinition,
			Name: &gqlast.Name{
				Kind:  gqlkind.Name,
				Value: prefix + s.Name.Name,
			},
			Description: &gqlast.StringValue{
				Kind:  gqlkind.StringValue,
				Value: gostrings.TrimSpace(s.Doc.Text()),
			},
			Fields: translateFields(t.Fields, translatorCfg),
		}
		doc.Definitions = append(doc.Definitions, def)
	}

	// print
	out := "# automatically generated file, do not edit!\n\n"
	out += gqlprinter.Print(doc).(string)
	f, err := os.Create(*output)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString(out); err != nil {
		log.Fatal(err)
	}
}

type translatorCfg struct {
	imports   map[string]string
	camelCase bool
	pkgPrefix bool
	pkgName   string
}

func translateFields(list *ast.FieldList, cfg *translatorCfg) []*gqlast.FieldDefinition {
	defs := make([]*gqlast.FieldDefinition, 0, len(list.List))
	for _, field := range list.List {
		// skip private fields
		if len(field.Names) > 0 && unicode.IsLower(rune(field.Names[0].Name[0])) {
			continue
		}
		// skip unmarshalled fields
		if gostrings.Contains(field.Tag.Value, `json:"-"`) {
			continue
		}
		defs = append(defs, translateField(field, cfg))
	}
	return defs
}

func getFieldName(f *ast.Field) (name string, snakecase bool) {
	if f.Tag != nil {
		re := regexp.MustCompile(`json:"(\w+),?.*"`)
		matches := re.FindSubmatch([]byte(f.Tag.Value))
		if len(matches) == 2 {
			return string(matches[1]), true
		}
	}
	if f.Names == nil || len(f.Names) == 0 {
		return nameFromExpr(f.Type), false
	}
	return f.Names[0].Name, false
}

func translateField(f *ast.Field, cfg *translatorCfg) *gqlast.FieldDefinition {
	// if camel case was requested convert field name
	name, snakeCase := getFieldName(f)
	if cfg.camelCase && snakeCase {
		name = underscoreToCamel(name)
	} else if cfg.camelCase && !snakeCase {
		name = pascalToCamel(name)
	}
	return &gqlast.FieldDefinition{
		Kind: gqlkind.FieldDefinition,
		Name: &gqlast.Name{
			Kind:  gqlkind.Name,
			Value: name,
		},
		Description: &gqlast.StringValue{
			Kind:  gqlkind.StringValue,
			Value: gostrings.TrimSpace(f.Doc.Text()),
		},
		Type: translateType(f.Type, false, cfg.pkgName, cfg),
	}
}

func nameFromExpr(t ast.Expr) string {
	switch v := t.(type) {
	case *ast.StarExpr:
		return nameFromExpr(v.X)
	case *ast.SelectorExpr:
		return nameFromExpr(v.Sel)
	case *ast.Ident:
		return v.Name
	}
	return ""
}

func translateType(t ast.Expr, opt bool, pkg string, cfg *translatorCfg) gqlast.Type {
	var val gqlast.Type
	switch v := t.(type) {
	case *ast.StarExpr:
		return translateType(v.X, true, pkg, cfg)
	case *ast.SelectorExpr:
		if t, ok := v.X.(*ast.Ident); ok && t.Name != "" {
			pkg = pathToPackageName(cfg.imports[t.Name])
		}
		return translateType(v.Sel, opt, pkg, cfg)
	case *ast.ArrayType:
		val = &gqlast.List{
			Kind: gqlkind.List,
			Type: translateType(v.Elt, false, pkg, cfg),
		}
	case *ast.Ident:
		name := v.Name
		switch name {
		case "string":
			name = "String"
		case "int64", "int32", "int":
			name = "Int"
		case "float64", "float32":
			name = "Float"
		case "bool":
			name = "Boolean"
		default:
			if name != "ObjectMeta" && cfg.pkgPrefix {
				name = pkg + name
			}
		}
		val = &gqlast.Named{
			Kind: gqlkind.Named,
			Name: &gqlast.Name{
				Kind:  gqlkind.Name,
				Value: name,
			},
		}
	default: // fallback to generic "JSON" type if no requisite type
		val = &gqlast.Named{
			Kind: gqlkind.Named,
			Name: &gqlast.Name{
				Kind:  gqlkind.Name,
				Value: "JSON",
			},
		}
	}
	if !opt {
		val = &gqlast.NonNull{
			Kind: gqlkind.NonNull,
			Type: val,
		}
	}
	return val
}

func discoverImports(path string) (map[string]string, error) {
	set := token.NewFileSet()
	packs, err := parser.ParseDir(set, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	out := map[string]string{}
	for _, pack := range packs {
		for _, file := range pack.Files {
			for _, i := range file.Imports {
				path := gostrings.Trim(i.Path.Value, "\"")
				name := ""
				if i.Name != nil {
					name = i.Name.Name
				}
				if name == "" {
					pathCmps := gostrings.Split(path, "/")
					name = pathCmps[len(pathCmps)-1]
				}
				out[name] = path
			}
		}
	}
	return out, nil
}

func discoverStructs(path string) ([]*ast.TypeSpec, error) {
	set := token.NewFileSet()
	packs, err := parser.ParseDir(set, path, nil, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	// indiana jones and the pyramid of doom
	out := []*ast.TypeSpec{}
	for _, pack := range packs {
		for _, file := range pack.Files {
			for _, d := range file.Decls {
				t, ok := d.(*ast.GenDecl)
				if !ok || t.Tok != token.TYPE {
					continue
				}
				for _, spec := range t.Specs {
					typeSpec, ok := spec.(*ast.TypeSpec)
					if !ok {
						continue
					}
					if _, isStruct := typeSpec.Type.(*ast.StructType); isStruct {
						typeSpec.Doc = t.Doc
						out = append(out, typeSpec)
					}
				}
			}
		}
	}

	return out, nil
}

// convert to package path to camel case; eg. core/v2 -> CoreV2
func pathToPackageName(path string) (prefix string) {
	segs := gostrings.Split(path, "/")
	segs = segs[2:] // snip: github.com/org
	if len(segs) == 1 {
		return gostrings.Title(segs[0])
	}

	// prefix is package name
	// unless it is versioned i.e. core/v2
	packageName := segs[len(segs)-1]
	prefix = gostrings.Title(packageName)
	if semver.IsValid(packageName) && len(segs) > 1 {
		prefix = gostrings.Title(segs[len(segs)-2]) + prefix
	}
	return
}

// https://github.com/asaskevich/govalidator/blob/3153c74/utils.go#L101
func underscoreToCamel(in string) string {
	head := in[:1]
	repl := gostrings.Replace(
		gostrings.Title(gostrings.Replace(gostrings.ToLower(in), "_", " ", -1)),
		" ",
		"",
		-1,
	)
	return head + repl[1:]
}

func pascalToCamel(in string) string {
	if len(in) == 0 {
		return in
	}
	head := in[:1]
	repl := gostrings.ToLower(head) + in[1:]
	return repl
}
