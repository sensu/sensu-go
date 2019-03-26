package v2

import (
	"context"
	"testing"
)

func TestPageSizeFromContext(t *testing.T) {
	tests := []struct {
		description string
		ctx         context.Context
		expected    int
	}{
		{
			description: "it returns 0 if there is no page size in the context",
			ctx:         context.Background(),
			expected:    0,
		},
		{
			description: "it returns the page size set in the context",
			ctx:         context.WithValue(context.Background(), PageSizeKey, 500),
			expected:    500,
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			got := PageSizeFromContext(test.ctx)

			if got != test.expected {
				t.Errorf("got %v, expected %v", got, test.expected)
			}
		})
	}
}

func TestPageContinueFromContext(t *testing.T) {
	tests := []struct {
		description string
		ctx         context.Context
		expected    string
	}{
		{
			description: "it returns an empty string if there is no continue token in the context",
			ctx:         context.Background(),
			expected:    "",
		},
		{
			description: "it returns the continue token set in the context",
			ctx:         context.WithValue(context.Background(), PageContinueKey, "sartre"),
			expected:    "sartre",
		},
	}

	for _, test := range tests {
		t.Run(test.description, func(t *testing.T) {
			got := PageContinueFromContext(test.ctx)

			if got != test.expected {
				t.Errorf("got %v, expected %v", got, test.expected)
			}
		})
	}
}
