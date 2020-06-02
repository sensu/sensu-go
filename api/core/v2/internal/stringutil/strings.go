package stringutil

import (
	"strings"
	"unicode"
)

const (
	lowerAlphaStart = 97
	lowerAlphaStop  = 122
)

// InArray searches 'array' for 'item' string
// Returns true 'item' is a value of 'array'
func InArray(item string, array []string) bool {
	if item == "" || len(array) == 0 {
		return false
	}

	for i := range array {
		if array[i] == item {
			return true
		}
	}

	return false
}

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

// Remove searches 'array' for 'item' string
// and removes the string if found
func Remove(item string, array []string) []string {
	for i, v := range array {
		if v == item {
			array = append(array[:i], array[i+1:]...)
			break
		}
	}
	return array
}

// Intersect finds the intersection between two slices of strings.
func Intersect(a []string, b []string) []string {
	set := make([]string, 0, len(a))
	tab := map[string]bool{}

	for _, e := range a {
		tab[e] = true
	}

	for _, e := range b {
		if _, found := tab[e]; found {
			set = append(set, e)
		}
	}
	return set
}
