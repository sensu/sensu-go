package stager

import (
	"context"
	"reflect"
	"sync"
	"testing"
)

func TestStager(t *testing.T) {
	var mx sync.Mutex
	var items []int
	st := New()

	s := st.NextStage()
	s.StartWithContext(func(ctx context.Context) {
		<-ctx.Done()
		mx.Lock()
		defer mx.Unlock()
		items = append(items, 1)
	})

	s = st.NextStage()
	s.StartWithContext(func(ctx context.Context) {
		<-ctx.Done()
		mx.Lock()
		defer mx.Unlock()
		items = append(items, 2)
	})

	s = st.NextStage()
	s.StartWithContext(func(ctx context.Context) {
		<-ctx.Done()
		mx.Lock()
		defer mx.Unlock()
		items = append(items, 3)
	})

	st.Shutdown()

	if !reflect.DeepEqual(items, []int{3, 2, 1}) {
		t.Errorf("unexpected result %v", items)
	}
}
