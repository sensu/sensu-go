package v2

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type contextTest[T any] struct {
	Value  T
	Create func(ctx context.Context, val T) context.Context
	Read   func(ctx context.Context) T
}

func (c contextTest[T]) Test(t *testing.T) {
	name := fmt.Sprintf("%T", c.Value)
	t.Run(name+"_success", func(t *testing.T) {
		ctx := context.Background()
		ctx = c.Create(ctx, c.Value)
		if got, want := c.Read(ctx), c.Value; !cmp.Equal(got, want) {
			t.Errorf("bad %T: got %v, want %v", got, got, want)
		}
	})
	t.Run(name+"_missing", func(t *testing.T) {
		value := c.Read(context.Background())
		if !reflect.ValueOf(value).IsZero() {
			t.Errorf("missing value is not zero value for %T: %v", value, value)
		}
	})
}

type tester interface {
	Test(*testing.T)
}

func TestContext(t *testing.T) {
	tests := []tester{
		contextTest[*TxInfo]{
			Value: &TxInfo{
				Records: []TxRecordInfo{{}},
			},
			Create: ContextWithTxInfo,
			Read:   TxInfoFromContext,
		},
		contextTest[IfMatch]{
			Value:  IfMatch{ETag("hello"), ETag("world")},
			Create: ContextWithIfMatch,
			Read:   IfMatchFromContext,
		},
		contextTest[IfNoneMatch]{
			Value:  IfNoneMatch{ETag("hello"), ETag("world")},
			Create: ContextWithIfNoneMatch,
			Read:   IfNoneMatchFromContext,
		},
	}
	for _, test := range tests {
		test.Test(t)
	}
}
