package generator

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/jamesdphillips/graphql/gqlerrors"
	"github.com/jamesdphillips/graphql/language/ast"
	"github.com/jamesdphillips/graphql/language/parser"
)

// GraphQLFileExt refers to extension used by GraphQL schema definition files.
const GraphQLFileExt = ".graphql"

// ParseDir parses all files in given path returning parsed AST and error if
// any occurred while parsing or given path is not valid.
func ParseDir(path string) (GraphQLFiles, error) {
	fs := GraphQLFiles{}
	fds, err := ioutil.ReadDir(path)
	if err != nil {
		return fs, err
	}

	for _, fd := range fds {
		if filepath.Ext(fd.Name()) != GraphQLFileExt {
			continue
		}
		f, err := ParseFile(filepath.Join(path, fd.Name()))
		if err != nil {
			return fs, err
		}
		fs = append(fs, f)
	}

	return fs, nil
}

// ParseFile parses given path to GraphQL file returning parsed AST and error if
// any occurred while parsing.
func ParseFile(path string) (*GraphQLFile, error) {
	bin, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	params := parser.ParseParams{Source: string(bin)}
	ast, err := parser.Parse(params)
	if err != nil {
		return nil, err
	}

	return &GraphQLFile{
		path: path,
		ast:  ast,
	}, nil
}

// GraphQLFiles encapsulates parsed document and filepath.
type GraphQLFiles []*GraphQLFile

// Dir returns path of where files can be found.
func (files GraphQLFiles) Dir() string {
	if len(files) == 0 {
		return ""
	}
	return filepath.Dir(files[0].path)
}

// Definitions returns all definition nodes found in document root.
func (files GraphQLFiles) Definitions() []ast.Node {
	defs := []ast.Node{}
	for _, f := range files {
		ds := f.Definitions()
		defs = append(defs, ds...)
	}
	return defs
}

// DefinitionsMap returns all definition nodes found in document root mapped by
// name.
func (files GraphQLFiles) DefinitionsMap() map[string]ast.Node {
	defs := make(map[string]ast.Node, len(files))
	for _, def := range files.Definitions() {
		name := getNodeName(def)
		defs[name] = def
	}
	return defs
}

// Validate returns an error if given files does not appear to describe a
// GraphQL Schema.
func (files GraphQLFiles) Validate() error {
	for _, f := range files {
		if err := f.Validate(); err != nil {
			return err
		}
	}

	defs := map[string]ast.Node{}
	for _, def := range files.Definitions() {
		name := getNodeName(def)
		if name == "" {
			continue
		}

		// Ensure there are no naming collisions
		if _, ok := defs[name]; ok {
			return newValidationErrorf(
				def,
				"node '%s' has already been defined",
				name,
			)
		}
	}

	return nil
}

func getNodeName(def ast.Node) string {
	switch d := def.(type) {
	case *ast.ScalarDefinition:
		return d.Name.Value
	case *ast.ObjectDefinition:
		return d.Name.Value
	case *ast.InputObjectDefinition:
		return d.Name.Value
	case *ast.InterfaceDefinition:
		return d.Name.Value
	case *ast.UnionDefinition:
		return d.Name.Value
	default:
		return ""
	}
}

// GraphQLFile encapsulates parsed document and filepath.
type GraphQLFile struct {
	path string
	ast  *ast.Document
}

// Filename returns the file's name
func (file *GraphQLFile) Filename() string {
	return filepath.Base(file.path)
}

// Definitions returns all definition nodes found in document root.
func (file *GraphQLFile) Definitions() []ast.Node {
	return file.ast.Definitions
}

// Validate returns an error if given file does not appear to be GraphQL Schema
// Definition Language (SDL).
func (file *GraphQLFile) Validate() error {
	defs := file.Definitions()
	for _, def := range defs { // TODO: Check more than top level?
		switch def.(type) {
		case *ast.OperationDefinition:
			return newValidationErrorf(def, "file should not define any operations")
		case *ast.FragmentDefinition:
			return newValidationErrorf(def, "file should not define any fragments")
		default:
			// TODO: Validate that unsupported directives are not given.
			// TODO: Validate names; eg. types should use upper CamelCase where as
			//       fields should use lower camelCase.
			continue
		}
	}
	return nil
}

func newValidationErrorf(node ast.Node, msg string, v ...interface{}) error {
	return gqlerrors.NewError(
		fmt.Sprintf(msg, v...),
		[]ast.Node{node},
		"",
		node.GetLoc().Source,
		[]int{},
		nil,
	)
}
