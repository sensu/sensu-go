package store

import (
	"fmt"
	"testing"
)

type BadStruct struct {
	Name        string
	BustedField int `hash:"string"`
}

type GoodStruct struct {
	Name string
}

type NestedStruct struct {
	Name  string
	Child GoodStruct
}

func TestETag(t *testing.T) {
	tests := []struct {
		name    string
		v       interface{}
		want    string
		wantErr bool
	}{
		{
			name:    "invalid resource",
			v:       &BadStruct{},
			wantErr: true,
		},
		{
			name:    "valid resource",
			v:       &GoodStruct{Name: "foo"},
			want:    `"860ef6ded7405308"`,
			wantErr: false,
		},
		{
			name:    "nested resource",
			v:       &NestedStruct{Name: "outer", Child: GoodStruct{Name: "inner"}},
			want:    `"2d733151c1aa0aed"`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ETag(tt.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("ETag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("ETag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestETag_Stability(t *testing.T) {
	v := &GoodStruct{Name: "foo"}
	etag1, err := ETag(v)
	if err != nil {
		t.Fatal(err)
	}

	etag2, err := ETag(v)
	if err != nil {
		t.Fatal(err)
	}

	if etag1 != etag2 {
		t.Errorf("ETag() is not stable: %s != %s", etag1, etag2)
	}
}

func TestCheckIfMatch(t *testing.T) {
	etag := `"abc"`
	tests := []struct {
		header string
		want   bool
	}{
		{"", true},
		{"*", true},
		{`"abc"`, true},
		{`"abc", "def"`, true},
		{`"xyz"`, false},
		{`W/"abc"`, false}, // Check that weak etag doesn't match (requires strong match)
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("header=%s", tt.header), func(t *testing.T) {
			if got := CheckIfMatch(tt.header, etag); got != tt.want {
				t.Errorf("CheckIfMatch(%q, %q) = %v, want %v", tt.header, etag, got, tt.want)
			}
		})
	}
}

func TestCheckIfNoneMatch(t *testing.T) {
	etag := `"abc"`
	tests := []struct {
		header string
		want   bool
	}{
		{"", true},
		{"*", false},
		{`"abc"`, false},
		{`W/"abc"`, false}, // Weak match should be successful for If-None-Match
		{`"xyz"`, true},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("header=%s", tt.header), func(t *testing.T) {
			if got := CheckIfNoneMatch(tt.header, etag); got != tt.want {
				t.Errorf("CheckIfNoneMatch(%q, %q) = %v, want %v", tt.header, etag, got, tt.want)
			}
		})
	}
}
