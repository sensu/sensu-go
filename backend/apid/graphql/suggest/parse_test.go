package suggest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseRef(t *testing.T) {
	testCases := []struct {
		in  string
		out RefComponents
		err error
	}{
		{
			in:  "core/v2/checks/subscriptions",
			out: RefComponents{Group: "core/v2", Name: "checks", FieldPath: "subscriptions"},
		},
		{
			in:  "core/v2/checks/labels/region",
			out: RefComponents{Group: "core/v2", Name: "checks", FieldPath: "labels/region"},
		},
		{
			in:  "core/v2/handlers",
			out: RefComponents{Group: "core/v2", Name: "handlers", FieldPath: ""},
		},
		{
			in:  "core",
			err: ErrInvalidRef,
		},
		{
			in:  "core/v2",
			err: ErrInvalidRef,
		},
		{
			in:  "core.v2.handlers.labels",
			err: ErrInvalidRef,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.in, func(t *testing.T) {
			c, err := ParseRef(tc.in)
			assert.EqualValues(t, c, tc.out)
			assert.EqualValues(t, err, tc.err)
		})
	}
}
