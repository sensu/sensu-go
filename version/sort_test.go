package version

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortByVersion(t *testing.T) {
	versions := []string{
		"2.0.0",
		"2.10.0",
		"2.3.0",
	}

	sort.Sort(byVersion(versions))

	assert.Equal(t, versions[0], "2.10.0")
	assert.Equal(t, versions[1], "2.3.0")
	assert.Equal(t, versions[2], "2.0.0")
}
