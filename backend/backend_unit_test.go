package backend

import (
	"errors"
	"testing"

	"go.uber.org/atomic"
)

var closedCt atomic.Int64

type mockStopper struct {
	hang   bool
	err    error
	closed int64
}

func (s *mockStopper) Stop() error {
	if s.hang {
		select {}
	}
	s.closed = closedCt.Inc()
	return s.err
}

func (s mockStopper) Name() string {
	return "hi"
}

func TestStopGroupNormalStop(t *testing.T) {
	var sg stopGroup
	stoppers := []*mockStopper{{}, {}, {}, {}}
	for _, s := range stoppers {
		sg.Add(s)
	}
	if err := sg.Stop(); err != nil {
		t.Fatal(err)
	}
	// validate stopper close order
	for i := 1; i < len(stoppers); i++ {
		if stoppers[i].closed != stoppers[i-1].closed-1 {
			t.Errorf("expected stopper %d to be closed before %d", i, i-1)
		}
	}
	// validate erorr handling
	sg.Add(&mockStopper{err: errors.New("err")})
	if err := sg.Stop(); err == nil {
		t.Fatal("expected non-nil error")
	}
}
