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
	"testing"
)

func TestCreateComment(t *testing.T) {
	c0 := newComment(startPosition, "")
	if got, want := len(c0.Lines), 1; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	c1 := newComment(startPosition, `hello
world`)
	if got, want := len(c1.Lines), 2; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	if got, want := c1.Lines[0], "hello"; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	if got, want := c1.Lines[1], "world"; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	if got, want := c1.Cstyle, true; got != want {
		t.Errorf("got [%v] want [%v]", c1, want)
	}
}

func TestTakeLastComment(t *testing.T) {
	c0 := newComment(startPosition, "hi")
	c1 := newComment(startPosition, "there")
	_, l := takeLastComment([]Visitee{c0, c1})
	if got, want := len(l), 1; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	if got, want := l[0], c0; got != want {
		t.Errorf("got [%v] want [%v]", c1, want)
	}
}

func TestParseCommentWithEmptyLinesAndTripleSlash(t *testing.T) {
	proto := `
// comment 1
// comment 2
//
// comment 3
/// comment 4`
	p := newParserOn(proto)
	def, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	//spew.Dump(def)
	if got, want := len(def.Elements), 1; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}

	if got, want := len(def.Elements[0].(*Comment).Lines), 5; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	if got, want := def.Elements[0].(*Comment).Lines[4], " comment 4"; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	if got, want := def.Elements[0].(*Comment).Position.Line, 2; got != want {
		t.Fatalf("got [%d] want [%d]", got, want)
	}
}

func TestParseCommentWithTripleSlash(t *testing.T) {
	proto := `
/// comment 1
`
	p := newParserOn(proto)
	def, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	//spew.Dump(def)
	if got, want := len(def.Elements), 1; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	if got, want := def.Elements[0].(*Comment).ExtraSlash, true; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	if got, want := def.Elements[0].(*Comment).Lines[0], " comment 1"; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	if got, want := def.Elements[0].(*Comment).Position.Line, 2; got != want {
		t.Fatalf("got [%d] want [%d]", got, want)
	}
}
