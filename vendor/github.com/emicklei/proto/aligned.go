// Copyright (c) 2017 Ernest Micklei
//
// MIT License
//
// Permission is hereby granted, free of charge, to any person obtaining
// a copy of this software and associated documentation files (the
// "Software"), to deal in the Software without restriction, including
// without limitation the rights to use, copy, modify, merge, publish,
// distribute, sublicense, and/or sell copies of the Software, and to
// permit persons to whom the Software is furnished to do so, subject to
// the following conditions:
//
// The above copyright notice and this permission notice shall be
// included in all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
// EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
// MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
// NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
// LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
// OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
// WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package proto

import "strings"
import "bytes"

type aligned struct {
	source  string
	left    bool
	padding bool
}

var (
	alignedEquals      = leftAligned(" = ")
	alignedShortEquals = leftAligned("=")
	alignedSpace       = leftAligned(" ")
	alignedComma       = leftAligned(", ")
	alignedEmpty       = leftAligned("")
	alignedSemicolon   = leftAligned(";")
)

func leftAligned(src string) aligned  { return aligned{src, true, true} }
func rightAligned(src string) aligned { return aligned{src, false, true} }
func notAligned(src string) aligned   { return aligned{src, false, false} }

func (a aligned) preferredWidth() int {
	if !a.hasAlignment() {
		return 0 // means do not force padding because of this source
	}
	return len(a.source)
}

func (a aligned) formatted(indentSeparator string, indentLevel, width int) string {
	if !a.padding {
		// if the source has newlines then make sure the correct indent level is applied
		buf := new(bytes.Buffer)
		for _, each := range a.source {
			buf.WriteRune(each)
			if '\n' == each {
				buf.WriteString(strings.Repeat(indentSeparator, indentLevel))
			}
		}
		return buf.String()
	}
	if a.left {
		return a.source + strings.Repeat(" ", width-len(a.source))
	}
	return strings.Repeat(" ", width-len(a.source)) + a.source
}

func (a aligned) hasAlignment() bool { return a.left || a.padding }
