package helpers

import (
	"bytes"
	"strings"
	"testing"

	corev2 "github.com/sensu/core/v2"
	"gopkg.in/yaml.v3"
)

// Verify PrintYAML emits 2-space indented YAML. yaml.v3 NewEncoder defaults
// to 4-space indent; enc.SetIndent(2) is required to match the prior v2 output.
// We check that the first child line after a top-level map key uses 2 spaces,
// not 4 — 4 spaces would be the regression symptom.
func TestPrintYAMLIndentation(t *testing.T) {
	check := corev2.FixtureCheckConfig("test")
	check.Namespace = "default"

	var buf bytes.Buffer
	if err := PrintYAML(check, &buf); err != nil {
		t.Fatal(err)
	}

	out := buf.String()
	lines := strings.Split(out, "\n")
	for i, line := range lines {
		// After a top-level key like "spec:" or "metadata:", the next
		// non-empty line should start with exactly 2 spaces, not 4.
		if (line == "spec:" || line == "metadata:") && i+1 < len(lines) {
			next := lines[i+1]
			// 4-space prefix without a leading 2-space-only line = wrong indent
			if strings.HasPrefix(next, "    ") {
				t.Errorf("4-space indent after %q — enc.SetIndent(2) not working:\n%s", line, out)
				return
			}
			if !strings.HasPrefix(next, "  ") {
				t.Errorf("expected 2-space indent after %q, got: %q\n%s", line, next, out)
				return
			}
		}
	}
}

// Verify multi-document output for a resource slice round-trips cleanly.
func TestPrintYAMLResourceSliceRoundTrip(t *testing.T) {
	checks := []corev2.CheckConfig{
		*corev2.FixtureCheckConfig("foo"),
		*corev2.FixtureCheckConfig("bar"),
	}

	var buf bytes.Buffer
	if err := PrintYAML(checks, &buf); err != nil {
		t.Fatal(err)
	}

	out := buf.String()

	// Round-trip: output must parse back to a slice with the same element count.
	var items []interface{}
	if err := yaml.Unmarshal([]byte(out), &items); err != nil {
		t.Fatalf("output is not valid YAML: %v\noutput:\n%s", err, out)
	}
	if len(items) != len(checks) {
		t.Errorf("expected %d items after round-trip, got %d\noutput:\n%s", len(checks), len(items), out)
	}
}
