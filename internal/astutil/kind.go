package astutil

import (
	"go/ast"
	"go/token"
	"sort"
)

const TypeMetaName = "TypeMeta"

// GetKindNames finds all the sensu-go kinds in a package. It returns a
// lexicographically sorted slice.
func GetKindNames(pkg *ast.Package) (result []string) {
	kinds := GetKinds(pkg)
	for k := range kinds {
		result = append(result, k)
	}
	sort.Strings(result)
	return result
}

// IsKind returns true if the type is an *ast.StructType, and it embeds
// meta.TypeMeta.
func IsKind(typ ast.Expr) bool {
	strukt, ok := typ.(*ast.StructType)
	if !ok {
		return false
	}
	for _, field := range strukt.Fields.List {
		if len(field.Names) != 0 {
			// not embedded
			continue
		}
		expr, ok := field.Type.(*ast.SelectorExpr)
		if !ok {
			continue
		}
		if expr.Sel.Name == TypeMetaName {
			return true
		}
	}
	return false
}

func GetKinds(pkg *ast.Package) map[string]*ast.TypeSpec {
	result := make(map[string]*ast.TypeSpec)
	for _, f := range pkg.Files {
		for _, decl := range f.Decls {
			gendecl, ok := decl.(*ast.GenDecl)
			if !ok || gendecl.Tok != token.TYPE {
				continue
			}
			for _, spec := range gendecl.Specs {
				ts, ok := spec.(*ast.TypeSpec)
				if !ok || !ts.Name.IsExported() {
					continue
				}
				if IsKind(ts.Type) {
					result[ts.Name.Name] = ts
				}
			}
		}
	}
	return result
}
