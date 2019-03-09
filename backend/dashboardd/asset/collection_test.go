package asset

import (
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewCollection(t *testing.T) {
	c := NewCollection()
	require.NotNil(t, c)

	// collection should usable even without any contents
	f, err := c.Open("/")
	assert.Nil(t, f)
	assert.Error(t, err) // not found
}

func TestCollectionExtend(t *testing.T) {
	c := NewCollection()
	require.NotNil(t, c)

	// should be able to extend without panic
	c.Extend(http.Dir("./fixtures/dir_a"))

	// collection should usable
	f, err := c.Open("/")
	assert.NotNil(t, f)
	assert.NoError(t, err)
}

func TestCollectionOpen(t *testing.T) {
	c := NewCollection()
	require.NotNil(t, c)

	// Extend with two FS
	c.Extend(http.Dir("./fixtures/dir_a"))
	c.Extend(http.Dir("./fixtures/dir_b"))

	// collection should list dir
	f, err := c.Open("/")
	assert.NotNil(t, f)
	assert.NoError(t, err)

	// collection should return file
	f, err = c.Open("/c")
	assert.NotNil(t, f)
	assert.NoError(t, err)

	// if two files match file in last FS should take precedence
	f, err = c.Open("/a")
	require.NoError(t, err)
	b, err := ioutil.ReadAll(f)
	require.NoError(t, err)
	assert.Equal(t, b[:1], []byte("2"))

	// not found
	f, err = c.Open("/this-file-should-definitely-not-exist")
	assert.Nil(t, f)
	assert.Error(t, err)
}
