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

func TestService(t *testing.T) {
	proto := `service AccountService {
		// comment
		rpc CreateAccount (CreateAccount) returns (ServiceFault);
		rpc GetAccounts   (stream Int64)  returns (Account);
	}`
	pr, err := newParserOn(proto).Parse()
	if err != nil {
		t.Fatal(err)
	}
	srv := collect(pr).Services()[0]
	if got, want := len(srv.Elements), 2; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	if got, want := srv.Position.String(), "<input>:1:1"; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	rpc1 := srv.Elements[0].(*RPC)
	if got, want := rpc1.Name, "CreateAccount"; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	if got, want := rpc1.Doc().Message(), " comment"; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	if got, want := rpc1.Position.Line, 3; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	rpc2 := srv.Elements[1].(*RPC)
	if got, want := rpc2.Name, "GetAccounts"; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
}

func TestRPCWithOptionAggregateSyntax(t *testing.T) {
	proto := `service AccountService {
		// CreateAccount
		rpc CreateAccount (CreateAccount) returns (ServiceFault){
			option (test_ident) = {
				test: "test"
				test2:"test2"
			};
		}
	}`
	pr, err := newParserOn(proto).Parse()
	if err != nil {
		t.Fatal(err)
	}
	srv := collect(pr).Services()[0]
	if got, want := len(srv.Elements), 1; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	rpc1 := srv.Elements[0].(*RPC)
	if got, want := len(rpc1.Options), 1; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	opt := rpc1.Options[0]
	if got, want := opt.Name, "(test_ident)"; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	if got, want := len(opt.AggregatedConstants), 2; got != want {
		t.Fatalf("got [%v] want [%v]", got, want)
	}
	if got, want := opt.AggregatedConstants[0].Source, "test"; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	if got, want := opt.AggregatedConstants[1].Source, "test2"; got != want {
		t.Errorf("got [%v] want [%v]", got, want)
	}
	t.Log(formatted(srv))
}
