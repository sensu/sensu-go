package backend

import (
	"errors"
	"testing"
)

type mockStopper struct {
	hang bool
	err  error
}

func (s mockStopper) Stop() error {
	if s.hang {
		select {}
	}
	return s.err
}

func (s mockStopper) Name() string {
	return "hi"
}

func TestStopGroupNormalStop(t *testing.T) {
	var sg stopGroup
	sg.Add(mockStopper{})
	sg.Add(mockStopper{})
	sg.Add(mockStopper{})
	if err := sg.Stop(); err != nil {
		t.Fatal(err)
	}
	sg.Add(mockStopper{err: errors.New("err")})
	if err := sg.Stop(); err == nil {
		t.Fatal("expected non-nil error")
	}
}
