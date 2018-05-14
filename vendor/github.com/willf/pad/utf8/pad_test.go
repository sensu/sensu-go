package utf8

import (
	"testing"
	"testing/quick"
	"unicode/utf8"
)

func TestLeftEqualWithSameLength(t *testing.T) {
	f := func(a string, pad string) bool {
		slen := utf8.RuneCountInString(a)
		padded := Left(a, slen, pad)
		return padded == a
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestRightEqualWithSameLength(t *testing.T) {
	f := func(a string, pad string) bool {
		slen := utf8.RuneCountInString(a)
		padded := Right(a, slen, pad)
		return padded == a
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestLeftEqualWithShorterLength(t *testing.T) {
	f := func(a string, pad string) bool {
		slen := utf8.RuneCountInString(a)
		padded := Left(a, slen-3, pad)
		return padded == a
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestRightEqualWithShorterLength(t *testing.T) {
	f := func(a string, pad string) bool {
		slen := utf8.RuneCountInString(a)
		padded := Right(a, slen-3, pad)
		return padded == a
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestLeftEqualWithGreaterLength(t *testing.T) {
	f := func(a string, pad string) bool {
		slen := utf8.RuneCountInString(a) + 3
		padded := Left(a, slen, pad)
		return padded == times(pad, 3)+a
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func TestRightEqualWithGreaterLength(t *testing.T) {
	f := func(a string, pad string) bool {
		slen := utf8.RuneCountInString(a) + 3
		padded := Right(a, slen, pad)
		return padded == a+times(pad, 3)
	}
	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}
