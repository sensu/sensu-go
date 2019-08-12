package testutil

import (
	"reflect"
	"testing"
)

// CompareError ensures that the provided err is expected
func CompareError(err error, expected bool, t *testing.T) {
	t.Helper()

	if expected && err == nil {
		t.Fatal("expected error, got none")
	} else if !expected && err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

// CompareErrorType ensures that the provided err is of the expected type.
// No comparison of error values is done.
func CompareErrorType(err error, expected reflect.Type, t *testing.T) {
	t.Helper()

	if expected != nil {
		if err == nil {
			t.Fatal("expected error, got none")
		}
		if reflect.TypeOf(err) != expected {
			t.Fatalf("expected error of type %s, got %T", expected.Name(), err)
		}
	} else if expected == nil && err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}
