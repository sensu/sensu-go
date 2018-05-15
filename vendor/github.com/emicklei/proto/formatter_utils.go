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

import (
	"fmt"
	"io"
)

// printDoc writes documentation is available
func (f *Formatter) printDoc(v Visitee) {
	if hasDoc, ok := v.(Documented); ok {
		if doc := hasDoc.Doc(); doc != nil {
			f.printComment(doc)
		}
	}
}

// printComment formats a Comment.
func (f *Formatter) printComment(c *Comment) {
	f.nl()
	if c.Cstyle {
		fmt.Fprintln(f.w, "/*")
	}
	for i, each := range c.Lines {
		f.indent(0)
		if c.Cstyle {
			// only skip first and last empty lines
			skip := (i == 0 && len(each) == 0) ||
				(i == len(c.Lines)-1 && len(each) == 0)
			if !skip {
				fmt.Fprintf(f.w, "%s\n", each)
			}
		} else {
			if c.ExtraSlash {
				fmt.Fprint(f.w, "/")
			}
			fmt.Fprintf(f.w, "//%s\n", each)
		}
	}
	if c.Cstyle {
		fmt.Fprintf(f.w, " */\n")
	}
}

// begin writes a newline if the last statement kind is different. always indents.
// if the Visitee has comment then print it.
func (f *Formatter) begin(stmt string, v Visitee) {
	// if not the first statement and different from last and on same indent level.
	if len(f.lastStmt) > 0 && f.lastStmt != stmt && f.lastLevel == f.indentLevel {
		f.nl()
	}
	f.lastStmt = stmt
	f.printDoc(v)
	f.indent(0)
}

func (f *Formatter) end(stmt string) {
	f.lastStmt = stmt
}

// indent changes the indent level and writes indentation.
func (f *Formatter) indent(diff int) {
	f.level(diff)
	for i := 0; i < f.indentLevel; i++ {
		io.WriteString(f.w, f.indentSeparator)
	}
}

// columnsPrintable is for elements that can be printed in aligned columns.
type columnsPrintable interface {
	columns() (cols []aligned)
	//doc() *Comment
}

func (f *Formatter) printListOfColumns(list []columnsPrintable) {
	if len(list) == 0 {
		return
	}
	// collect all column values
	values := [][]aligned{}
	widths := map[int]int{}
	for _, each := range list {
		cols := each.columns()
		values = append(values, cols)
		// update max widths per column
		for i, other := range cols {
			pw := other.preferredWidth()
			w, ok := widths[i]
			if ok {
				if pw > w {
					widths[i] = pw
				}
			} else {
				widths[i] = pw
			}
		}
	}
	// now print all values
	for _, each := range values {
		f.indent(0)
		for c := 0; c < len(widths); c++ {
			pw := widths[c]
			// only print if there is a value
			if c < len(each) {
				// using space padding to match the max width
				io.WriteString(f.w, each[c].formatted(f.indentSeparator, f.indentLevel, pw))
			}
		}
		f.nl()
	}
}

// nl writes a newline.
func (f *Formatter) nl() {
	io.WriteString(f.w, "\n")
}

// level changes the current indentLevel
func (f *Formatter) level(diff int) {
	f.lastLevel = f.indentLevel
	f.indentLevel += diff
}

// printAsGroups prints the list in groups of the same element type.
func (f *Formatter) printAsGroups(list []Visitee) {
	group := []columnsPrintable{}
	lastGroupName := ""
	for _, each := range list {
		groupName := nameOfVisitee(each)
		printable, isColumnsPrintable := each.(columnsPrintable)
		if isColumnsPrintable {
			if lastGroupName != groupName {
				lastGroupName = groupName
				// print current group
				if len(group) > 0 {
					f.printListOfColumns(group)
					// begin new group
					group = []columnsPrintable{}
				}
			}
			// comment as a group entity
			if hasDoc, ok := each.(Documented); ok {
				if doc := hasDoc.Doc(); doc != nil {
					f.printListOfColumns(group)
					// begin new group
					group = append([]columnsPrintable{}, doc.columnsPrintables()...)
				}
			}
			group = append(group, printable)
		} else {
			// not printable in group
			lastGroupName = groupName
			// print current group
			if len(group) > 0 {
				f.printListOfColumns(group)
				// begin new group
				group = []columnsPrintable{}
			}
			each.Accept(f)
		}
	}
	// print last group
	f.printListOfColumns(group)
}

// endWithComment writes a statement end (;) followed by inline comment if present.
func (f *Formatter) endWithComment(commentOrNil *Comment) {
	io.WriteString(f.w, ";")
	if commentOrNil != nil {
		if commentOrNil.ExtraSlash {
			io.WriteString(f.w, " ///")
		} else {
			io.WriteString(f.w, " //")
		}
		io.WriteString(f.w, commentOrNil.Message())
	}
	io.WriteString(f.w, "\n")
}
