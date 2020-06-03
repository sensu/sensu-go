package main

import "testing"

func TestSnakeCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "single uppercase character",
			input: "A",
			want:  "a",
		},
		{
			name:  "single-word with some uppercases",
			input: "FooBar",
			want:  "foo_bar",
		},
		{
			name:  "all uppercase word",
			input: "FOO",
			want:  "foo",
		},
		{
			name:  "lowercase with underscore",
			input: "foo_bar",
			want:  "foo_bar",
		},
		{
			name:  "some uppercases with underscore",
			input: "Foo_Bar",
			want:  "foo_bar",
		},
		{
			name:  "acronym",
			input: "TLSOptions",
			want:  "tls_options",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := snakeCase(tt.input); got != tt.want {
				t.Errorf("snakeCase() = %v, want %v", got, tt.want)
			}
		})
	}
}
