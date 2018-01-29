package leader

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWork(t *testing.T) {
	f := func(context.Context) error {
		return nil
	}
	w := newWork(f)
	assert.NotNil(t, w.f)
	assert.NotNil(t, w.result)
	w2 := newWork(f)
	assert.True(t, w.id < w2.id)
	err := errors.New("foo")
	w2.result <- err
	assert.Equal(t, err, w2.Err())
	// returns the same result after multiple invocations
	assert.Equal(t, err, w2.Err())
}
