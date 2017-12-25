package integration

import (
	"bytes"
	"go/parser"
	"go/token"
	"testing"

	"github.com/dave/jennifer/jen"
	"github.com/sensu/sensu-go/backend/apid/graphql/generator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestSaver struct {
	out string
}

func (t *TestSaver) Save(f *jen.File) error {
	buf := &bytes.Buffer{}
	if err := f.Render(buf); err != nil {
		return err
	}
	t.out = buf.String()
	return nil
}

func TestKitchenSinkExample(t *testing.T) {
	file, err := generator.ParseFile("./schema-kitchen-sink.graphql")
	require.NoError(t, err)
	require.NotNil(t, file)
	require.NoError(t, file.Validate())

	generator := generator.New("mypackage", file)
	require.NotNil(t, generator)

	saver := TestSaver{}
	generator.Saver = &saver

	gerr := generator.Run()
	require.NoError(t, gerr)
	assert.NotEmpty(t, saver.out)

	perr := parseSrc(saver.out)
	assert.NoError(t, perr)
}

func parseSrc(src string) error {
	fset := token.NewFileSet()
	_, err := parser.ParseFile(fset, "", src, parser.AllErrors)
	return err
}
