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

import "testing"

func TestEnum(t *testing.T) {
	proto := `
// enum
enum EnumAllowingAlias {
  option allow_alias = true;
  UNKNOWN = 0;
  STARTED = 1;
  RUNNING = 2 [(custom_option) = "hello world"];
  NEG = -42;
}`
	p := newParserOn(proto)
	pr, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	enums := collect(pr).Enums()
	if got, want := len(enums), 1; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	if got, want := len(enums[0].Elements), 5; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	if got, want := enums[0].Comment != nil, true; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	if got, want := enums[0].Comment.Message(), " enum"; got != want {
		t.Errorf("got [%v] want [%v]", enums[0].Comment, want)
	}
	if got, want := enums[0].Position.Line, 3; got != want {
		t.Errorf("got [%d] want [%d]", got, want)
	}
	ef1 := enums[0].Elements[1].(*EnumField)
	if got, want := ef1.Integer, 0; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	if got, want := ef1.Position.Line, 5; got != want {
		t.Errorf("got [%d] want [%d]", got, want)
	}
	ef3 := enums[0].Elements[3].(*EnumField)
	if got, want := ef3.Integer, 2; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	if got, want := ef3.ValueOption.Name, "(custom_option)"; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	if got, want := ef3.ValueOption.Constant.Source, "hello world"; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	if got, want := ef3.Position.Line, 7; got != want {
		t.Errorf("got [%d] want [%d]", got, want)
	}
	ef4 := enums[0].Elements[4].(*EnumField)
	if got, want := ef4.Integer, -42; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
}
