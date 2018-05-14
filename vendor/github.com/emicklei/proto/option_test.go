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

func TestOptionCases(t *testing.T) {
	for i, each := range []struct {
		proto     string
		name      string
		strLit    string
		nonStrLit string
	}{{
		`option (full).java_package = "com.example.foo";`,
		"(full).java_package",
		"com.example.foo",
		"",
	}, {
		`option Bool = true;`,
		"Bool",
		"",
		"true",
	}, {
		`option Float = -3.14E1;`,
		"Float",
		"",
		"-3.14E1",
	}, {
		`option (foo_options) = { opt1: 123 opt2: "baz" };`,
		"(foo_options)",
		"",
		"",
	}, {
		`option optimize_for = SPEED;`,
		"optimize_for",
		"",
		"SPEED",
	}} {
		p := newParserOn(each.proto)
		pr, err := p.Parse()
		if err != nil {
			t.Fatal("testcase failed:", i, err)
		}
		if got, want := len(pr.Elements), 1; got != want {
			t.Fatalf("[%d] got [%v] want [%v]", i, got, want)
		}
		o := pr.Elements[0].(*Option)
		if got, want := o.Name, each.name; got != want {
			t.Errorf("[%d] got [%v] want [%v]", i, got, want)
		}
		if len(each.strLit) > 0 {
			if got, want := o.Constant.Source, each.strLit; got != want {
				t.Errorf("[%d] got [%v] want [%v]", i, got, want)
			}
		}
		if len(each.nonStrLit) > 0 {
			if got, want := o.Constant.Source, each.nonStrLit; got != want {
				t.Errorf("[%d] got [%v] want [%v]", i, got, want)
			}
		}
		if got, want := o.IsEmbedded, false; got != want {
			t.Errorf("[%d] got [%v] want [%v]", i, got, want)
		}
	}
}

func TestLiteralString(t *testing.T) {
	proto := `"string"`
	p := newParserOn(proto)
	l := new(Literal)
	if err := l.parse(p); err != nil {
		t.Fatal(err)
	}
	if got, want := l.IsString, true; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	if got, want := l.Source, "string"; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
}

func TestOptionComments(t *testing.T) {
	proto := `
// comment
option Help = "me"; // inline`
	p := newParserOn(proto)
	pr, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	o := pr.Elements[0].(*Option)
	if got, want := o.IsEmbedded, false; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	if got, want := o.Comment != nil, true; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	if got, want := o.Comment.Lines[0], " comment"; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	if got, want := o.InlineComment != nil, true; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	if got, want := o.InlineComment.Lines[0], " inline"; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	if got, want := o.Position.Line, 3; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	if got, want := o.Comment.Position.Line, 2; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	if got, want := o.InlineComment.Position.Line, 3; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
}

func TestIssue8(t *testing.T) {
	proto := `
// usage:
message Bar {
  // alternative aggregate syntax (uses TextFormat):
  int32 b = 2 [(foo_options) = {
    opt1: 123,
    opt2: "baz"
  }];
}
	`
	p := newParserOn(proto)
	pr, err := p.Parse()
	if err != nil {
		t.Fatal(err)
	}
	o := pr.Elements[0].(*Message)
	f := o.Elements[0].(*NormalField)
	if got, want := len(f.Options), 1; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	ac := f.Options[0].AggregatedConstants
	if got, want := len(ac), 2; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	if got, want := ac[0].Name, "opt1"; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	if got, want := ac[1].Name, "opt2"; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	if got, want := ac[0].Source, "123"; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	if got, want := ac[1].Source, "baz"; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	if got, want := o.Position.Line, 3; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	if got, want := o.Comment.Position.String(), "<input>:2:1"; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	if got, want := f.Position.String(), "<input>:5:3"; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	if got, want := ac[0].Position.Line, 6; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	if got, want := ac[1].Position.Line, 7; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
}
