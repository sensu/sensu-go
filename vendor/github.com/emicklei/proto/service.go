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
	"io"
	"text/scanner"
)

// Service defines a set of RPC calls.
type Service struct {
	Position scanner.Position
	Comment  *Comment
	Name     string
	Elements []Visitee
}

// Accept dispatches the call to the visitor.
func (s *Service) Accept(v Visitor) {
	v.VisitService(s)
}

// Doc is part of Documented
func (s *Service) Doc() *Comment {
	return s.Comment
}

// addElement is part of elementContainer
func (s *Service) addElement(v Visitee) {
	s.Elements = append(s.Elements, v)
}

// elements is part of elementContainer
func (s *Service) elements() []Visitee {
	return s.Elements
}

// takeLastComment is part of elementContainer
// removes and returns the last elements of the list if it is a Comment.
func (s *Service) takeLastComment() (last *Comment) {
	last, s.Elements = takeLastComment(s.Elements)
	return
}

// parse continues after reading "service"
func (s *Service) parse(p *Parser) error {
	pos, tok, lit := p.next()
	if tok != tIDENT {
		if !isKeyword(tok) {
			return p.unexpected(lit, "service identifier", s)
		}
	}
	s.Name = lit
	pos, tok, lit = p.next()
	if tok != tLEFTCURLY {
		return p.unexpected(lit, "service opening {", s)
	}
	for {
		pos, tok, lit = p.next()
		switch tok {
		case tCOMMENT:
			if com := mergeOrReturnComment(s.Elements, lit, pos); com != nil { // not merged?
				s.Elements = append(s.Elements, com)
			}
		case tRPC:
			rpc := new(RPC)
			rpc.Position = pos
			rpc.Comment, s.Elements = takeLastComment(s.Elements)
			err := rpc.parse(p)
			if err != nil {
				return err
			}
			s.Elements = append(s.Elements, rpc)
		case tSEMICOLON:
			maybeScanInlineComment(p, s)
		case tRIGHTCURLY:
			goto done
		default:
			return p.unexpected(lit, "service comment|rpc", s)
		}
	}
done:
	return nil
}

// RPC represents an rpc entry in a message.
type RPC struct {
	Position       scanner.Position
	Comment        *Comment
	Name           string
	RequestType    string
	StreamsRequest bool
	ReturnsType    string
	StreamsReturns bool
	Options        []*Option
	InlineComment  *Comment
}

// Accept dispatches the call to the visitor.
func (r *RPC) Accept(v Visitor) {
	v.VisitRPC(r)
}

// Doc is part of Documented
func (r *RPC) Doc() *Comment {
	return r.Comment
}

// inlineComment is part of commentInliner.
func (r *RPC) inlineComment(c *Comment) {
	r.InlineComment = c
}

// columns returns printable source tokens
func (r *RPC) columns() (cols []aligned) {
	cols = append(cols,
		leftAligned("rpc "),
		leftAligned(r.Name),
		leftAligned(" ("))
	if r.StreamsRequest {
		cols = append(cols, leftAligned("stream "))
	} else {
		cols = append(cols, alignedEmpty)
	}
	cols = append(cols,
		leftAligned(r.RequestType),
		leftAligned(") "),
		leftAligned("returns"),
		leftAligned(" ("))
	if r.StreamsReturns {
		cols = append(cols, leftAligned("stream "))
	} else {
		cols = append(cols, alignedEmpty)
	}
	cols = append(cols,
		leftAligned(r.ReturnsType),
		leftAligned(")"))
	if len(r.Options) > 0 {
		buf := new(bytes.Buffer)
		io.WriteString(buf, " {\n")
		f := NewFormatter(buf, "  ") // TODO get separator, now 2 spaces
		f.level(1)
		for _, each := range r.Options {
			each.Accept(f)
			io.WriteString(buf, "\n")
		}
		f.indent(-1)
		io.WriteString(buf, "}")
		cols = append(cols, notAligned(buf.String()))
	} else {
		cols = append(cols, alignedSemicolon)
	}
	if r.InlineComment != nil {
		cols = append(cols, notAligned(" //"), notAligned(r.InlineComment.Message()))
	}
	return cols
}

// parse continues after reading "rpc"
func (r *RPC) parse(p *Parser) error {
	pos, tok, lit := p.next()
	if tok != tIDENT {
		return p.unexpected(lit, "rpc method", r)
	}
	r.Name = lit
	pos, tok, lit = p.next()
	if tok != tLEFTPAREN {
		return p.unexpected(lit, "rpc type opening (", r)
	}
	pos, tok, lit = p.next()
	if tSTREAM == tok {
		r.StreamsRequest = true
		pos, tok, lit = p.next()
	}
	if tok != tIDENT {
		return p.unexpected(lit, "rpc stream | request type", r)
	}
	r.RequestType = lit
	pos, tok, lit = p.next()
	if tok != tRIGHTPAREN {
		return p.unexpected(lit, "rpc type closing )", r)
	}
	pos, tok, lit = p.next()
	if tok != tRETURNS {
		return p.unexpected(lit, "rpc returns", r)
	}
	pos, tok, lit = p.next()
	if tok != tLEFTPAREN {
		return p.unexpected(lit, "rpc type opening (", r)
	}
	pos, tok, lit = p.next()
	if tSTREAM == tok {
		r.StreamsReturns = true
		pos, tok, lit = p.next()
	}
	if tok != tIDENT {
		return p.unexpected(lit, "rpc stream | returns type", r)
	}
	r.ReturnsType = lit
	pos, tok, lit = p.next()
	if tok != tRIGHTPAREN {
		return p.unexpected(lit, "rpc type closing )", r)
	}
	pos, tok, lit = p.next()
	if tSEMICOLON == tok {
		p.nextPut(pos, tok, lit) // allow for inline comment parsing
		return nil
	}
	if tLEFTCURLY == tok {
		// parse options
		for {
			pos, tok, lit = p.next()
			if tRIGHTCURLY == tok {
				break
			}
			if tCOMMENT == tok {
				// TODO put comment in the next option
				continue
			}
			if tOPTION != tok {
				return p.unexpected(lit, "rpc option", r)
			}
			o := new(Option)
			o.Position = pos
			if err := o.parse(p); err != nil {
				return err
			}
			r.Options = append(r.Options, o)
		}
	}
	return nil
}
