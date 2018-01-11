package generator

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/dave/jennifer/jen"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/location"
)

// GeneratedFileExt describes the default extension used when the resulting Go
// file is written.
const GeneratedFileExt = ".gql.go"

// DefaultPackageName refers to default name given to resulting package.
const DefaultPackageName = "schema"

// invoker is used to when generating comments to inform reader that what
// package / app / script generated the given code.
const defaultInvoker = "graphql/generator package"

// Saver represents an item that can write generated types.
type Saver interface {
	Save(name string, f *jen.File) error
}

// DryRun implements Saver interface, omits writing code to disk and logs
// result of render.
type DryRun struct {
	Debug bool
}

// Save runs render and logs results
func (d *DryRun) Save(name string, f *jen.File) error {
	buf := &bytes.Buffer{}
	if err := f.Render(buf); err != nil {
		logger.WithError(err).Error("unable to render generated code")
		return err
	}
	logger.WithField("file", name).Info("dry-run successful; rendered without err")
	if d.Debug {
		logger.Print(buf)
	}
	return nil
}

// fileSaver writes generated types to disk given path.
type fileSaver struct {
	sourceDir string
}

// Save renders types to disk
func (s fileSaver) Save(fname string, f *jen.File) error {
	buf := &bytes.Buffer{}
	if err := f.Render(buf); err != nil {
		return err
	}
	outpath := makeOutputPath(s.sourceDir, fname, GeneratedFileExt)
	return ioutil.WriteFile(outpath, buf.Bytes(), 0644)
}

// File extension .gql.go is used for generated files
func makeOutputPath(dir, name, newExt string) string {
	ext := filepath.Ext(name)
	fpath := name[0 : len(name)-len(ext)]
	return filepath.Join(dir, fpath+newExt)
}

// Generator generates Go code for type defnitions found in given source files.
type Generator struct {
	// Saver handles rendering and persisting generators output.
	Saver

	// Invoker field identifies the caller that invoked generated. Name is
	// included in warning comment at top of generated file.
	Invoker string

	// PackageName of given to resulting files. Defaults to "schema."
	PackageName string

	source GraphQLFiles
}

// New returns new generator given path and name of package resulting file will
// reside.
func New(source GraphQLFiles) Generator {
	return Generator{
		Saver:       fileSaver{sourceDir: source.Dir()},
		Invoker:     defaultInvoker,
		PackageName: DefaultPackageName,
		source:      source,
	}
}

// Run generates code and saves
func (g Generator) Run() error {
	// Wrap contextual information about current step
	i := newInfo(g.source)

	// Generate code for each source file
	outfiles := make(map[string]*jen.File, len(g.source))
	for _, s := range g.source {
		outfile := newFile(g.PackageName, g.Invoker)
		generateCode(s, i, outfile)
		outfiles[s.Filename()] = outfile
	}

	// Do dry run to ensure that the files can be written.
	for _, outfile := range outfiles {
		if err := outfile.Render(ioutil.Discard); err != nil {
			return err
		}
	}

	// Write generated code to disk.
	for name, outfile := range outfiles {
		if err := g.Save(name, outfile); err != nil {
			return err
		}
	}

	return nil
}

func generateCode(source *GraphQLFile, i info, file *jen.File) {
	// Iterate through each definition found in the document and generate
	// appropriate code.
	for _, d := range source.Definitions() {
		// Update contextual information about current iteration
		i := withUpdatedInfo(i, source, getNodeName(d))

		// Assuming code was generated for node append to file
		if code := genTypeDefinition(d, i); code != nil {
			file.Add(code)
		}
	}
}

func genTypeDefinition(node ast.Node, i info) jen.Code {
	loc := location.GetLocation(node.GetLoc().Source, node.GetLoc().Start)
	logger := logger.WithField("type", node.GetKind()).WithField("line", loc.Line)

	switch def := node.(type) {
	case *ast.ScalarDefinition:
		return genScalar(def)
	case *ast.ObjectDefinition:
		return genObjectType(def, i)
	case *ast.UnionDefinition:
		return genUnion(def)
	case *ast.EnumDefinition:
		return genEnum(def)
	case *ast.InterfaceDefinition:
		return genInterface(def)
	case *ast.InputObjectDefinition:
		return genInputObject(def)
	case *ast.SchemaDefinition:
		return genSchema(def)
	case *ast.DirectiveDefinition:
		logger.Warn("unsupported at this time; skipping")
	case *ast.TypeExtensionDefinition:
		logger.Warn("unsupported at this time; skipping")
	default:
		logger.Fatal("unhandled type encountered")
	}
	return nil
}

// Used when generating code to lookup adjacent definitions, provide contextual
// information.
type info struct {
	files       GraphQLFiles
	definitions map[string]ast.Node
	currentFile *GraphQLFile
	currentNode string
}

func newInfo(files GraphQLFiles) info {
	return info{
		files:       files,
		definitions: files.DefinitionsMap(),
	}
}

func withUpdatedInfo(i info, file *GraphQLFile, node string) info {
	i.currentFile = file
	i.currentNode = node
	return i
}

func newFile(name, invoker string) *jen.File {
	// New file abstract w/ package name
	file := jen.NewFile(name)

	// Warning comment
	file.HeaderComment(fmt.Sprintf("Code generated by %s. DO NOT EDIT.", invoker))
	file.Line()

	return file
}
