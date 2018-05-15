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

func TestMessage(t *testing.T) {
	proto := `
		message   Out   {
		// identifier
		string   id  = 1;
		// size
		int64   size = 2;

		oneof foo {
			string     name        = 4;
			SubMessage sub_message = 9;
		}
		message  Inner {   // Level 2
   			int64  ival = 1;
  		}
		map<string, testdata.SubDefaults> proto2_value  =  13;
		option  (my_option).a  =  true;
	}`
	p := newParserOn(proto)
	p.next() // consume first token
	m := new(Message)
	err := m.parse(p)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := m.Name, "Out"; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	if got, want := len(m.Elements), 6; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	if got, want := m.Elements[0].(*NormalField).Position.String(), "<input>:4:3"; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	if got, want := m.Elements[0].(*NormalField).Comment.Position.String(), "<input>:3:3"; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	if got, want := m.Elements[3].(*Message).Position.String(), "<input>:12:3"; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	if got, want := m.Elements[3].(*Message).Elements[0].(*NormalField).Position.Line, 13; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
}
