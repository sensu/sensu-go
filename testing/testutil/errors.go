package testutil

import "testing"

// CompareError ensures that the provided err is expected
func CompareError(err error, expected bool, t *testing.T) {
	t.Helper()

	if expected && err == nil {
		t.Fatal("expected error, got none")
	} else if !expected && err != nil {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}
