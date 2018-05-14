package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type (
	StringMap map[string]string
)

func (s *StringMap) String() string {
	pairs := make([]string, 0, len(*s))

	for key, value := range *s {
		pairs = append(pairs, key+"="+value)
	}

	return strings.Join(pairs, ",")
}

func (s *StringMap) Set(value string) error {
	parts := strings.Split(value, ",")

	for _, part := range parts {
		pair := strings.Split(part, "=")

		(*s)[strings.TrimSpace(pair[0])] = strings.TrimSpace(pair[1])
	}

	return nil
}

var (
	stdOut bool

	txtOut string

	goOut string

	jsOut string

	resolveImports StringMap

	packageAliases StringMap

	noPrefix bool
)

func main() {
	resolveImports = make(StringMap)
	packageAliases = make(StringMap)

	flag.BoolVar(&stdOut, "std_out", false, "Writes transformed files to stdout")
	flag.StringVar(&txtOut, "txt_out", "", "Writes transformed files to .graphql file")
	flag.StringVar(&goOut, "go_out", "", "Writes transformed files to .go file")
	flag.StringVar(&jsOut, "js_out", "", "Writes transformed files to .js file")
	flag.Var(&resolveImports, "resolve_import", "Resolves given external packages")
	flag.Var(&packageAliases, "package_alias", "Renames packages using given aliases")
	flag.BoolVar(&noPrefix, "no_prefix", false, "Disables package prefix for type names")

	flag.Parse()

	if len(flag.Args()) == 0 {
		flag.Usage()
		os.Exit(0)
	}

	var transformer *Transformer
	writers := make([]io.Writer, 0, 5)

	if stdOut == true {
		writers = append(writers, os.Stdout)
	}

	if txtOut != "" {
		writer, err := createTextWriter(txtOut)

		if err != nil {
			gracefullyTerminate(err, writers)
		}

		writers = append(writers, writer)
	}

	if goOut != "" {
		writer, err := createGoWriter(goOut)

		if err != nil {
			gracefullyTerminate(err, writers)
		}

		writers = append(writers, writer)
	}

	if jsOut != "" {
		writer, err := createJsWriter(jsOut)

		if err != nil {
			gracefullyTerminate(err, writers)
		}

		writers = append(writers, writer)
	}

	if len(writers) == 0 {
		fmt.Println("output not defined")
		os.Exit(0)
	}

	transformer = NewTransformer(
		io.MultiWriter(writers...),
		withResolvingImports(resolveImports),
		withPackageAliases(packageAliases),
		withNoPrefix(noPrefix),
	)

	var err error

	for _, filename := range flag.Args() {
		if err = readAndTransform(filename, transformer); err != nil {
			fmt.Println(fmt.Sprintf("failed to transform %s file : %s", filename, err.Error()))

			break
		}
	}

	err = saveWriters(writers)

	if err != nil {
		os.Exit(1)
	}
}

func withResolvingImports(imports StringMap) func(transformer *Transformer) {
	return func(t *Transformer) {
		for key, url := range imports {
			t.Import(key, url)
		}
	}
}

func withPackageAliases(aliases StringMap) func(transformer *Transformer) {
	return func(t *Transformer) {
		for pkg, alias := range aliases {
			t.SetPackageAlias(pkg, alias)
		}
	}
}

func withNoPrefix(noPrefix bool) func(transformer *Transformer) {
	return func(t *Transformer) {
		t.DisablePrefix(noPrefix)
	}
}

func ensureExtension(filename, expectedExt string) string {
	ext := filepath.Ext(filename)

	if ext == "" {
		return filename + expectedExt
	}

	if ext != expectedExt {
		return strings.Replace(filename, ext, expectedExt, -1)
	}

	return filename
}

func createTextWriter(filename string) (io.Writer, error) {
	return NewFileWriter(ensureExtension(filename, ".graphql"), "", "")
}

func createGoWriter(filename string) (io.Writer, error) {
	filename = ensureExtension(filename, ".go")
	abs, err := filepath.Abs(filename)

	if err != nil {
		return nil, err
	}

	name := strings.Replace(filepath.Base(abs), ".go", "", -1)

	openTag := "package " + filepath.Base(filepath.Dir(abs)) + "\n \n"
	openTag += "var " + strings.Title(name) + " = `\n"

	return NewFileWriter(filename, openTag, "\n`")
}

func createJsWriter(filename string) (io.Writer, error) {
	openTag := "module.exports = `\n"

	return NewFileWriter(ensureExtension(filename, ".js"), openTag, "\n`")
}

func saveWriters(writers []io.Writer) error {
	var err error

	for _, writer := range writers {
		fw, ok := writer.(*FileWriter)

		if ok == true {
			if err == nil {
				err = fw.Save()

				if err != nil {
					fmt.Println(err.Error())
				}
			} else {
				fw.Remove()
			}
		}
	}

	return err
}

func gracefullyTerminate(err error, writers []io.Writer) {
	for _, writer := range writers {
		fw, ok := writer.(*FileWriter)

		if ok == true {
			fw.Remove()
		}
	}

	fmt.Println("error occured: " + err.Error())
	os.Exit(1)
}

func readAndTransform(filename string, transformer *Transformer) error {
	// open for read
	file, err := os.Open(filename)

	if err != nil {
		return err
	}

	defer file.Close()

	if err := transformer.Transform(file); err != nil {
		return err
	}

	return nil
}
