/*
Package utf8 provides left-padding and right-padding for utf8 strings


*/
package utf8

import "unicode/utf8"

func times(str string, n int) (out string) {
	for i := 0; i < n; i++ {
		out += str
	}
	return
}

// Left left-pads the string with pad up to len runes
// len may be exceeded if
func Left(str string, len int, pad string) string {
	return times(pad, len-utf8.RuneCountInString(str)) + str
}

// Right right-pads the string with pad up to len runes
func Right(str string, len int, pad string) string {
	return str + times(pad, len-utf8.RuneCountInString(str))
}
