package validators

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateTrueFalse(t *testing.T) {
	tests := []struct {
		Input    interface{}
		ExpError bool
	}{
		{
			Input:    0,
			ExpError: true,
		},
		{
			Input:    "yesh",
			ExpError: true,
		},
		{
			Input:    "nein",
			ExpError: true,
		},
		{
			Input: "y",
		},
		{
			Input: "Y",
		},
		{
			Input: "Yes",
		},
		{
			Input: "YeS",
		},
		{
			Input: "n",
		},
		{
			Input: "N",
		},
		{
			Input: "no",
		},
		{
			Input: "No",
		},
		{
			Input: "NO",
		},
		{
			Input: "t",
		},
		{
			Input: "T",
		},
		{
			Input: "True",
		},
		{
			Input: "true",
		},
		{
			Input: "f",
		},
		{
			Input: "false",
		},
		{
			Input: "False",
		},
	}
	for _, test := range tests {
		err := ValidateTrueFalse(test.Input)
		f := assert.Nil
		if test.ExpError {
			f = assert.NotNil
		}
		f(t, err)
	}
}
