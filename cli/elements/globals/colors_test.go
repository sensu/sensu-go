package globals

import (
	"testing"

	"github.com/mgutz/ansi"
	"github.com/stretchr/testify/assert"
)

func TestBooleanStyle(t *testing.T) {
	assert := assert.New(t)

	// changes true
	trueIn := "true"
	trueOut := BooleanStyle(trueIn)
	assert.NotEqual(trueIn, trueOut)
	assert.Contains(trueOut, ansi.ColorCode("blue"))

	// changes false
	falseIn := "false"
	falseOut := BooleanStyle(falseIn)
	assert.NotEqual(falseIn, falseOut)
	assert.Contains(falseOut, ansi.ColorCode("red"))

	// neither 'true' or 'false'
	neitherIn := "neither lol"
	neitherOut := BooleanStyle(neitherIn)
	assert.Equal(neitherIn, neitherOut)
	assert.NotContains(neitherOut, ansi.ColorCode("red"))
	assert.NotContains(neitherOut, ansi.ColorCode("blue"))
}

func TestBooleanStyleP(t *testing.T) {
	assert := assert.New(t)

	trueOut := BooleanStyleP(true)
	assert.Contains(trueOut, "true")

	falseOut := BooleanStyleP(false)
	assert.Contains(falseOut, "false")
}
