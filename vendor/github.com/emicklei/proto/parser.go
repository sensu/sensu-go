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
	"errors"
	"fmt"
	"io"
	"runtime"
	"strconv"
	"text/scanner"
)

var startPosition = scanner.Position{Line: 1, Column: 1}

// Parser represents a parser.
type Parser struct {
	debug         bool
	scanner       *scanner.Scanner
	buf           *nextValues
	scannerErrors []error
}

// nextValues is to capture the result of next()
type nextValues struct {
	pos scanner.Position
	tok token
	lit string
}

// NewParser returns a new instance of Parser.
func NewParser(r io.Reader) *Parser {
	s := new(scanner.Scanner)
	s.Init(r)
	s.Mode = scanner.ScanIdents | scanner.ScanFloats | scanner.ScanStrings | scanner.ScanRawStrings | scanner.ScanComments
	p := &Parser{scanner: s}
	s.Error = p.handleScanError
	return p
}

// handleScanError is called from the underlying Scanner
func (p *Parser) handleScanError(s *scanner.Scanner, msg string) {
	p.scannerErrors = append(p.scannerErrors,
		fmt.Errorf("go scanner error at %v = %v", s.Position, msg))
}

// Parse parses a proto definition. May return a parse or scanner error.
func (p *Parser) Parse() (*Proto, error) {
	proto := new(Proto)
	parseError := proto.parse(p)
	// see if it was a scanner error
	if len(p.scannerErrors) > 0 {
		buf := new(bytes.Buffer)
		for _, each := range p.scannerErrors {
			fmt.Fprintln(buf, each)
		}
		return proto, errors.New(buf.String())
	}
	return proto, parseError
}

// Filename is for reporting. Optional.
func (p *Parser) Filename(f string) {
	p.scanner.Filename = f
}

// next returns the next token using the scanner or drain the buffer.
func (p *Parser) next() (pos scanner.Position, tok token, lit string) {
	if p.buf != nil {
		// consume buf
		vals := *p.buf
		p.buf = nil
		return vals.pos, vals.tok, vals.lit
	}
	ch := p.scanner.Scan()
	if ch == scanner.EOF {
		return p.scanner.Position, tEOF, ""
	}
	lit = p.scanner.TokenText()
	return p.scanner.Position, asToken(lit), lit
}

// nextPut sets the buffer
func (p *Parser) nextPut(pos scanner.Position, tok token, lit string) {
	p.buf = &nextValues{pos, tok, lit}
}

func (p *Parser) unexpected(found, expected string, obj interface{}) error {
	debug := ""
	if p.debug {
		_, file, line, _ := runtime.Caller(1)
		debug = fmt.Sprintf(" at %s:%d (with %#v)", file, line, obj)
	}
	return fmt.Errorf("found %q on %v, expected [%s]%s", found, p.scanner.Position, expected, debug)
}

func (p *Parser) nextInteger() (i int, err error) {
	_, tok, lit := p.next()
	if "-" == lit {
		i, err = p.nextInteger()
		return i * -1, err
	}
	if tok != tIDENT {
		return 0, errors.New("non integer")
	}
	i, err = strconv.Atoi(lit)
	return
}

// nextIdentifier consumes tokens which may have one or more dot separators (namespaced idents).
func (p *Parser) nextIdentifier() (pos scanner.Position, tok token, lit string) {
	pos, tok, lit = p.next()
	if tIDENT != tok {
		return
	}
	startPos := pos
	fullLit := lit
	// see if identifier is namespaced
	for {
		r := p.scanner.Peek()
		if '.' != r {
			break
		}
		p.next() // consume dot
		pos, tok, lit := p.next()
		if tIDENT != tok {
			p.nextPut(pos, tok, lit)
			break
		}
		fullLit = fmt.Sprintf("%s.%s", fullLit, lit)
	}
	return startPos, tIDENT, fullLit
}

func (p *Parser) peekNonWhitespace() rune {
	r := p.scanner.Peek()
	if r == scanner.EOF {
		return r
	}
	if isWhitespace(r) {
		// consume it
		p.scanner.Next()
		return p.peekNonWhitespace()
	}
	return r
}
