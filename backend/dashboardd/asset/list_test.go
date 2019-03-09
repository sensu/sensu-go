package asset

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListContents(t *testing.T) {
	// list all
	fs := http.Dir("./fixtures/dir_b")
	ls, err := ListContents(fs, "/")
	require.NoError(t, err)
	require.NotEmpty(t, ls)
	assert.Len(t, ls, 2)

	// sub directories are expanded
	fs = http.Dir("./fixtures/dir_a")
	ls, err = ListContents(fs, "/")
	require.NoError(t, err)
	require.NotEmpty(t, ls)
	assert.Len(t, ls, 5)

	// given bad path
	ls, err = ListContents(fs, "/bad-path")
	assert.Error(t, err)
	assert.Empty(t, ls)

	// given file
	ls, err = ListContents(fs, "/d/e")
	assert.Empty(t, ls)
	assert.Error(t, err)
}
