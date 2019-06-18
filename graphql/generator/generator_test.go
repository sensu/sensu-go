package generator

import (
	"go/parser"
	"go/token"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratorRun(t *testing.T) {
	fs, err := ParseDir("./kitchen-sink-schema")
	require.NoError(t, err)

	generator := New(fs)
	require.NotNil(t, generator)

	saver := testSaver{}
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
