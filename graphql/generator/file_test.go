package generator

import (
	"bytes"
	"testing"

	"github.com/dave/jennifer/jen"
	"github.com/stretchr/testify/require"
)

type testSaver struct {
	out string
}

func (t *testSaver) Save(_ string, f *jen.File) error {
	buf := &bytes.Buffer{}
	if err := f.Render(buf); err != nil {
		return err
	}
	t.out = buf.String()
	return nil
}

func TestLoadDir(t *testing.T) {
	fs, err := ParseDir("./kitchen-sink-schema")
	require.NoError(t, err)
	require.NotEmpty(t, fs)
	require.NoError(t, fs.Validate())
}

func TestFilesValidate(t *testing.T) {
	fs, _ := ParseDir("./kitchen-sink-schema")
	err := fs.Validate()
	require.NoError(t, err)
}

func TestFileValidate(t *testing.T) {
	testCases := []struct {
		desc      string
		path      string
		expectErr bool
	}{
		{
			desc:      "given valid schema",
			path:      "./kitchen-sink-schema/schema.graphql",
			expectErr: false,
		},
		{
			desc:      "given graphql file with query",
			path:      "./schema-with-query.graphql",
			expectErr: true,
		},
		{
			desc:      "given graphql file with fragment",
			path:      "./schema-with-fragment.graphql",
			expectErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.desc, func(t *testing.T) {
			f, _ := ParseFile(tc.path)
			err := f.Validate()
			if tc.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
