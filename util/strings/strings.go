package strings

import (
	"regexp"
	"strings"
)

// InArray searches 'array' for 'item' string
// Returns true 'item' is a value of 'array'
func InArray(item string, array []string) bool {
	if item == "" || len(array) == 0 {
		return false
	}

	for _, element := range array {
		if element == item {
			return true
		}
	}

	return false
}

// FoundInArray searches array for item without distinguishing between uppercase
// and lowercase and non-alphanumeric characters. Returns true if item is a
// value of array
func FoundInArray(item string, array []string) bool {
	if item == "" || len(array) == 0 {
		return false
	}

	// Prepare our regex in order to remove all non-alphanumeric characters
	r, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		return false
	}

	item = r.ReplaceAllString(strings.ToLower(item), "")

	for _, element := range array {
		element = r.ReplaceAllString(strings.ToLower(element), "")
		if element == item {
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
