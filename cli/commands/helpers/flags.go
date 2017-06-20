package helpers

import (
	"fmt"
	"regexp"
	"strings"
)

var commaWhitespaceRegex *regexp.Regexp

// SafeSplitCSV splits given string and trims and extraneous whitespace
func SafeSplitCSV(i string) []string {
	trimmed := strings.TrimSpace(i)
	trimmed = commaWhitespaceRegex.ReplaceAllString(trimmed, ",")

	if len(trimmed) > 0 {
		return strings.Split(trimmed, ",")
	}

	return []string{}
}

func init() {
	// Matches same whitespace that the stdlib's unicode or strings packages would
	// https://golang.org/src/unicode/graphic.go?s=3997:4022#L116
	whiteSpc := "\\t\\n\\v\\f\\r\u0085\u00A0 "
	commaWhitespaceRegex = regexp.MustCompile(
		fmt.Sprintf("[%s]*,[%s]*", whiteSpc, whiteSpc),
	)
}
