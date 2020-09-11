package etcd

import (
	"testing"
)

func Test_checkIfMatch(t *testing.T) {
	tests := []struct {
		name   string
		header string
		etag   string
		want   bool
	}{
		{
			name: "empty header should return true",
			etag: `"12345"`,
			want: true,
		},
		{
			name:   "empty etag should return false",
			header: `""`,
			etag:   `"12345"`,
			want:   false,
		},
		{
			name:   "the asterisk is a special value that represents any resource",
			header: "*",
			etag:   `"12345"`,
			want:   true,
		},
		{
			name:   "matching etag",
			header: `"12345"`,
			etag:   `"12345"`,
			want:   true,
		},
		{
			name:   "multiple etags can be passed in the header",
			header: `"000", "12345"`,
			etag:   `"12345"`,
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkIfMatch(tt.header, tt.etag); got != tt.want {
				t.Errorf("checkIfMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_checkIfNoneMatch(t *testing.T) {
	tests := []struct {
		name   string
		header string
		etag   string
		want   bool
	}{
		{
			name: "empty header should return true",
			etag: `"12345"`,
			want: true,
		},
		{
			name:   "empty etag should return true",
			header: `""`,
			etag:   `"12345"`,
			want:   true,
		},
		{
			name:   "the asterisk is a special value that represents any resource",
			header: "*",
			etag:   `"12345"`,
			want:   false,
		},
		{
			name:   "matching etag",
			header: `"12345"`,
			etag:   `"12345"`,
			want:   false,
		},
		{
			name:   "multiple etags can be passed in the header",
			header: `"000", "12345"`,
			etag:   `"12345"`,
			want:   false,
		},
		{
			name:   "the W/ prefix is handled",
			header: `W/"12345"`,
			etag:   `"12345"`,
			want:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := checkIfNoneMatch(tt.header, tt.etag); got != tt.want {
				t.Errorf("checkIfNoneMatch() = %v, want %v", got, tt.want)
			}
		})
	}
}
