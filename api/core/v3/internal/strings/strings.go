package strings

import (
	"strings"
	"unicode"
)

const (
	lowerAlphaStart = 97
	lowerAlphaStop  = 122
)

func isAlpha(r rune) bool {
	return r >= lowerAlphaStart && r <= lowerAlphaStop
}

func alphaNumeric(s string) bool {
	for _, r := range s {
		if !(unicode.IsDigit(r) || isAlpha(r)) {
			return false
		}
	}
	return true
}

func normalize(s string) string {
	if alphaNumeric(s) {
		return s
	}
	lowered := strings.ToLower(s)
	if alphaNumeric(lowered) {
		return lowered
	}
	trimmed := make([]rune, 0, len(lowered))
	for _, r := range lowered {
		if isAlpha(r) {
			trimmed = append(trimmed, r)
		}
	}
	return string(trimmed)
}

// FoundInArray searches array for item without distinguishing between uppercase
// and lowercase and non-alphanumeric characters. Returns true if item is a
// value of array
func FoundInArray(item string, array []string) bool {
	if item == "" || len(array) == 0 {
		return false
	}

	item = normalize(item)

	for i := range array {
		if normalize(array[i]) == item {
			return true
		}
	}

	return false
}
