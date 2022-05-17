package selector

import (
	"fmt"
	"io"
	"strings"
	"unicode"
)

// TokenT represents the type of lexer tokens
type TokenT int

type Token struct {
	Type  TokenT
	Value string
}

const (
	// start represents the starting state
	start TokenT = iota

	// endOfStringToken represents the end of the input string
	endOfStringToken

	// errorToken represents an error while tokenizing the input string
	errorToken

	// leftSquareToken represents (
	leftSquareToken

	// rightSquareToken represents )
	rightSquareToken

	// doubleEqualSignToken reprensents ==
	doubleEqualSignToken

	// notEqualToken represents !=
	notEqualToken

	// doubleAmpersandToken represents &&
	doubleAmpersandToken

	// inToken represents in
	inToken

	// notInToken represents notin
	notInToken

	// commaToken represents ,
	commaToken

	// identifierToken represents a strings of letters and/or digits
	identifierToken

	// stringToken represents a quoted string
	stringToken

	// boolToken represents a boolean (true, false)
	boolToken

	// matchesToken represents matches
	matchesToken
)

var reservedWords = map[string]Token{
	"in":      Token{Type: inToken, Value: "in"},
	"notin":   Token{Type: notInToken, Value: "notin"},
	"true":    Token{Type: boolToken, Value: "true"},
	"false":   Token{Type: boolToken, Value: "false"},
	"matches": Token{Type: matchesToken, Value: "matches"},
}

func newLexer(input string) *lexer {
	return &lexer{input: strings.NewReader(input)}
}

// lexer tokenises the input string
type lexer struct {
	input    io.RuneScanner
	position int
}

func identStart(r rune) bool {
	return r == '_' || unicode.IsLetter(r)
}

func identTail(r rune) bool {
	return r == '_' || r == '.' || r == '/' || unicode.IsDigit(r) || unicode.IsLetter(r)
}

// Tokenize returns the next token found in the input stream
func (l *lexer) Tokenize() Token {
	// Yep, it's a state machine. 1-rune lookahead.
	var state TokenT
	var buf []rune
	for {
		r, size, err := l.input.ReadRune()
		if err != nil && err != io.EOF {
			return Token{Type: errorToken, Value: err.Error()}
		}
		l.position += size
		if err == io.EOF {
			switch state {
			case start:
				if len(buf) > 0 {
					return Token{Type: errorToken, Value: fmt.Sprintf("end of input while scanning identifier: %q", string(buf))}
				}
				return Token{Type: endOfStringToken}
			case identifierToken:
			default:
				return Token{Type: errorToken}
			}
		}
		switch state {
		case start:
			if unicode.IsSpace(r) {
				continue
			}
			switch r {
			case '[':
				return Token{Type: leftSquareToken, Value: "["}
			case ']':
				return Token{Type: rightSquareToken, Value: "]"}
			case '=':
				state = doubleEqualSignToken
				buf = append(buf, r)
			case '!':
				state = notEqualToken
				buf = append(buf, r)
			case '&':
				state = doubleAmpersandToken
				buf = append(buf, r)
			case ',':
				return Token{Type: commaToken, Value: ","}
			case '"', '\'':
				state = stringToken
			default:
				if !identStart(r) {
					if len(buf) > 0 {
						return Token{Type: errorToken, Value: fmt.Sprintf("invalid identifier: %q", string(append(buf, r)))}
					}
					return Token{Type: errorToken, Value: fmt.Sprintf("invalid rune: %q", string(append(buf, r)))}
				}
				state = identifierToken
				buf = append(buf, r)
			}
		case doubleEqualSignToken, notEqualToken:
			switch r {
			case '=':
				return Token{Type: state, Value: string(append(buf, r))}
			default:
				errmsg := fmt.Sprintf("at %d, looking for %q but got %q", l.position, "=", string(append(buf, r)))
				return Token{Type: errorToken, Value: errmsg}
			}
		case doubleAmpersandToken:
			switch r {
			case '&':
				return Token{Type: state, Value: string(append(buf, r))}
			default:
				errmsg := fmt.Sprintf("at %d, looking for %q but got %q", l.position, "&", string(append(buf, r)))
				return Token{Type: errorToken, Value: errmsg}
			}
		case identifierToken:
			if unicode.IsSpace(r) || err == io.EOF {
				if buf[len(buf)-1] == '.' {
					return Token{Type: errorToken, Value: fmt.Sprintf("invalid identifier: %q", string(buf))}
				}
				buf := string(buf)
				lbuf := strings.ToLower(buf)
				if tok, ok := reservedWords[lbuf]; ok {
					return tok
				}
				return Token{Type: state, Value: buf}
			}
			switch r {
			case '[', ']', '!', '&', ',', '"', '\'':
				_ = l.input.UnreadRune()
				return Token{Type: state, Value: string(buf)}
			case '.':
				state = start
			}
			if !identTail(r) {
				return Token{Type: errorToken, Value: fmt.Sprintf("invalid identifier: %q", string(append(buf, r)))}
			}
			buf = append(buf, r)
		case stringToken:
			if err == io.EOF {
				return Token{Type: errorToken, Value: "EOF while scanning string literal"}
			}
			switch r {
			case '"', '\'':
				return Token{Type: state, Value: string(buf)}
			default:
				buf = append(buf, r)
			}
		}
	}
}
