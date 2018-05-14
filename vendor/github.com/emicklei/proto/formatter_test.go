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
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestPrintListOfColumns(t *testing.T) {
	e0 := new(EnumField)
	e0.Name = "A"
	e0.Integer = 1
	op0 := new(Option)
	op0.IsEmbedded = true
	op0.Name = "a"
	op0.Constant = Literal{Source: "1234"}
	e0.ValueOption = op0

	e1 := new(EnumField)
	e1.Name = "ABC"
	e1.Integer = 12
	op1 := new(Option)
	op1.IsEmbedded = true
	op1.Name = "ab"
	op1.Constant = Literal{Source: "1234"}
	e1.ValueOption = op1

	list := []columnsPrintable{e0, e1}
	b := new(bytes.Buffer)
	f := NewFormatter(b, " ")
	f.printListOfColumns(list)
	formatted := `A   =  1 [a  = 1234];
ABC = 12 [ab = 1234];
`
	if got, want := b.String(), formatted; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
}

func TestFormatCStyleComment(t *testing.T) {
	t.Skip()
	proto := `/*
 * Hello
 * World
 */
`
	def, _ := NewParser(strings.NewReader(proto)).Parse()
	b := new(bytes.Buffer)
	f := NewFormatter(b, " ")
	f.Format(def)
	if got, want := proto, formatted(def.Elements[0]); got != want {
		println(diff(got, want))
		t.Fail()
	}
}

func TestFormatExtendMessage(t *testing.T) {
	t.Skip()
	proto := `
// extend
extend google.protobuf.MessageOptions {
  // my_option
  optional string my_option = 51234; // mynumber
}
`
	p := newParserOn(proto)
	pp, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	m, ok := pp.Elements[0].(*Message)
	if !ok {
		t.Fatal("message expected")
	}
	if got, want := formatted(m), proto; got != want {
		fmt.Println(diff(got, want))
		fmt.Println(got)
		t.Fail()
	}
}

func TestFormatAggregatedOptionSyntax(t *testing.T) {
	// TODO format not that nice
	proto := `rpc Find (Finder) returns (stream Result) {
  option (google.api.http) = {
    post: "/v1/finders/1"
    body: "*"
  };

}
`
	p := newParserOn(proto)
	r := new(RPC)
	p.next() // consumer rpc
	err := r.parse(p)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := formatted(r), proto; got != want {
		fmt.Println(diff(got, want))
		fmt.Println("---")
		fmt.Println(got)
		fmt.Println("---")
		fmt.Println(want)
		t.Fail()
	}
}

func TestFormatCommentSample(t *testing.T) {
	proto := `
/*
 begin
*/

// comment 1
// comment 2
syntax = "proto"; // inline 1

// comment 3
// comment 4
package test; // inline 2

// comment 5
// comment 6
message Test {
    // comment 7
    // comment 8
    int64 i = 1; // inline 3
}	
/// triple
`
	p := newParserOn(proto)
	def, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	if got, want := len(def.Elements), 5; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	b := new(bytes.Buffer)
	f := NewFormatter(b, "  ") // 2 spaces
	f.Format(def)
	//println(b.String())
	//spew.Dump(def)
}

/// testing utils
func formatted(v Visitee) string {
	b := new(bytes.Buffer)
	f := NewFormatter(b, "  ") // 2 spaces
	v.Accept(f)
	return b.String()
}

func diff(left, right string) string {
	b := new(bytes.Buffer)
	w := func(char rune) {
		if '\n' == char {
			b.WriteString("(n)")
		} else if '\t' == char {
			b.WriteString("(t)")
		} else if ' ' == char {
			b.WriteString("( )")
		} else {
			b.WriteRune(char)
		}
	}
	b.WriteString("got:\n")
	for _, char := range left {
		w(char)
	}
	if len(left) == 0 {
		b.WriteString("(empty)")
	}
	b.WriteString("\n")
	for _, char := range right {
		w(char)
	}
	b.WriteString("\n:wanted\n")
	return b.String()
}
