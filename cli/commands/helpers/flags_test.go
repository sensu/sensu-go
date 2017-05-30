package helpers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafeSplitCSV(t *testing.T) {
	assert := assert.New(t)

	res := SafeSplitCSV("one")
	assert.Equal(res, []string{"one"})

	res = SafeSplitCSV("one,two")
	assert.Equal(res, []string{"one", "two"})

	res = SafeSplitCSV("one   ,       two")
	assert.Equal(res, []string{"one", "two"})

	res = SafeSplitCSV("one,     \t \u00a0 two")
	assert.Equal(res, []string{"one", "two"})

	res = SafeSplitCSV("    one ,     \t 🐛 two")
	assert.Equal(res, []string{"one", "🐛 two"})
}
