package v2

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/mitchellh/hashstructure"
)

// ETag represents a unique hash of the resource.
type ETag []byte

// String returns the base64-encoded unquoted form of the ETag.
func (e ETag) String() string {
	return base64.RawStdEncoding.EncodeToString(e)
}

// DecodeETag attempts to parse an unquoted encoded ETag. It returns an error if the
// input is not base64-encoded.
func DecodeETag(data string) (ETag, error) {
	b, err := base64.RawStdEncoding.DecodeString(data)
	return ETag(b), err
}

func (e ETag) Equals(other ETag) bool {
	return bytes.Equal(e, other)
}

// ETagFromStruct is for types that don't want to compute etag in the database
func ETagFromStruct(v interface{}) ETag {
	hash, err := hashstructure.Hash(v, nil)
	if err != nil {
		panic(err)
	}
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, hash)
	return ETag(b)
}

// IfMatch is a list of Etags
type IfMatch []ETag
type contextKeyIfMatch struct{}

func (m IfMatch) Matches(etag ETag) bool {
	for _, candidate := range m {
		if bytes.Equal(candidate, etag) {
			return true
		}
	}
	return false
}

func (m IfMatch) String() string {
	parts := make([]string, 0, len(m))
	for _, v := range m {
		parts = append(parts, fmt.Sprintf("%q", v.String()))
	}
	return strings.Join(parts, ", ")
}

func ReadIfMatch(header string) (IfMatch, error) {
	var result IfMatch
	parts := strings.Split(header, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if len(part) < 2 {
			continue
		}
		// remove quotes
		part = part[1 : len(part)-1]
		etag, err := DecodeETag(part)
		if err != nil {
			continue
		}
		result = append(result, etag)
	}
	return result, nil
}

// IfNoneMatch is a list of ETags.
type IfNoneMatch []ETag
type contextKeyIfNoneMatch struct{}

func (m IfNoneMatch) String() string {
	parts := make([]string, 0, len(m))
	for _, v := range m {
		parts = append(parts, fmt.Sprintf("%q", v.String()))
	}
	return strings.Join(parts, ", ")
}

func ReadIfNoneMatch(header string) (IfNoneMatch, error) {
	m, err := ReadIfMatch(header)
	return IfNoneMatch(m), err
}

func (m IfNoneMatch) Matches(etag ETag) bool {
	for _, candidate := range m {
		if bytes.Equal(candidate, etag) {
			return false
		}
	}
	return true
}
