package store

import (
	"fmt"
	"net/textproto"
	"strings"

	"github.com/mitchellh/hashstructure"
)

// ETag returns a unique hash for the given interface
func ETag(v interface{}) (string, error) {
	hash, err := hashstructure.Hash(v, nil)
	if err != nil {
		return "", err
	}
	hex := fmt.Sprintf("%x", hash)
	return fmt.Sprintf("%q", hex), nil
}

// ETagCondition represents a conditional request
type ETagCondition struct {
	IfMatch     string
	IfNoneMatch string
}

// CheckIfMatch determines if any of the etag provided in the If-Match header
// match the stored etag. This function was largely inspired by the net/http
// package
func CheckIfMatch(header string, etag string) bool {
	if header == "" {
		return true
	}

	for {
		header = textproto.TrimString(header)
		if len(header) == 0 {
			break
		}
		if header[0] == ',' {
			header = header[1:]
			continue
		}
		if header[0] == '*' {
			return true
		}
		scannedEtag, remainingHeader := scanETag(header)
		if scannedEtag == etag && scannedEtag != "" && scannedEtag[0] == '"' {
			return true
		}
		header = remainingHeader
	}

	return false
}

// CheckIfNoneMatch determines if none of the etag provided in the If-Match
// header match the stored etag. This function was largely inspired by the
// net/http package
func CheckIfNoneMatch(header string, etag string) bool {
	if header == "" {
		return true
	}

	for {
		header = textproto.TrimString(header)
		if len(header) == 0 {
			break
		}
		if header[0] == ',' {
			header = header[1:]
			continue
		}
		if header[0] == '*' {
			return false
		}
		scannedEtag, remainingHeader := scanETag(header)
		if strings.TrimPrefix(scannedEtag, "W/") == strings.TrimPrefix(etag, "W/") {
			return false
		}
		header = remainingHeader
	}

	return true
}

func scanETag(header string) (string, string) {
	header = textproto.TrimString(header)
	start := 0
	if strings.HasPrefix(header, "W/") {
		start = 2
	}

	if len(header[start:]) < 2 || header[start] != '"' {
		return "", ""
	}

	// ETag is either W/"text" or "text".
	// See RFC 7232 2.3.
	for i := start + 1; i < len(header); i++ {
		c := header[i]
		switch {
		// Character values allowed in ETags.
		case c == 0x21 || c >= 0x23 && c <= 0x7E || c >= 0x80:
		case c == '"':
			return string(header[:i+1]), header[i+1:]
		default:
			break
		}
	}

	return "", ""
}
