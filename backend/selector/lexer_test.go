package selector

import (
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLexerIgnoreSpaces(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Token
	}{
		{
			name:  "no space",
			input: "a",
			want:  Token{Type: identifierToken, Value: "a"},
		},
		{
			name:  "single whitespace",
			input: " a",
			want:  Token{Type: identifierToken, Value: "a"},
		},
		{
			name:  "two whitespaces",
			input: "  a",
			want:  Token{Type: identifierToken, Value: "a"},
		},
		{
			name:  "tabulation",
			input: "\ta",
			want:  Token{Type: identifierToken, Value: "a"},
		},
		{
			name:  "carriage",
			input: "\ra",
			want:  Token{Type: identifierToken, Value: "a"},
		},
		{
			name:  "new line",
			input: "\na",
			want:  Token{Type: identifierToken, Value: "a"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &lexer{
				input: strings.NewReader(tt.input),
			}
			if got := l.Tokenize(); got != tt.want {
				t.Errorf("error at %d: lexer.Token() = %v, want %v", l.position, got, tt.want)
			}
		})
	}
}

func TestLexerScanIdentifier(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Token
	}{
		{
			name:  "word",
			input: "foo",
			want:  Token{Type: identifierToken, Value: "foo"},
		},
		{
			name:  "sentence",
			input: "foo bar",
			want:  Token{Type: identifierToken, Value: "foo"},
		},
		{
			name:  "literal operator",
			input: "in ",
			want:  Token{Type: inToken, Value: "in"},
		},
		{
			name:  "not a literal operator",
			input: "inbefore",
			want:  Token{Type: identifierToken, Value: "inbefore"},
		},
		{
			name:  "literal operator with uppercase",
			input: "In ",
			want:  Token{Type: inToken, Value: "in"},
		},
		{
			name:  "string with special character",
			input: "'us-west-2:foo:*'",
			want:  Token{Type: stringToken, Value: "us-west-2:foo:*"},
		},
		{
			name:  "matches operator",
			input: "matches ",
			want:  Token{Type: matchesToken, Value: "matches"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &lexer{
				input: strings.NewReader(tt.input),
			}
			if got := l.Tokenize(); got != tt.want {
				t.Errorf("lexer.Token() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLexerScanOperator(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		want         Token
		wantPosition int
	}{
		{
			name:         "single character operator",
			input:        "[",
			want:         Token{Type: leftSquareToken, Value: "["},
			wantPosition: 1,
		},
		{
			name:         "two characters operator",
			input:        "==",
			want:         Token{Type: doubleEqualSignToken, Value: "=="},
			wantPosition: 2,
		},
		{
			name:         "single character operators with special character",
			input:        "[!=",
			want:         Token{Type: leftSquareToken, Value: "["},
			wantPosition: 1,
		},
		{
			name:         "operator with identifier",
			input:        "!= foo",
			want:         Token{Type: notEqualToken, Value: "!="},
			wantPosition: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &lexer{
				input: strings.NewReader(tt.input),
			}
			if got := l.Tokenize(); got != tt.want {
				t.Errorf("lexer.Tokenize() got = %v, want %v", got, tt.want)
			}
			assert.Equal(t, tt.wantPosition, l.position)
		})
	}
}

func TestLexerTokenize(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Token
	}{
		{
			name:  "identifier",
			input: "foo",
			want:  Token{Type: identifierToken, Value: "foo"},
		},
		{
			name:  "literal operator",
			input: "in foo",
			want:  Token{Type: inToken, Value: "in"},
		},
		{
			name:  "non-literal operator",
			input: "[foo]",
			want:  Token{Type: leftSquareToken, Value: "["},
		},
		{
			name:  "comparison operator",
			input: "!=",
			want:  Token{Type: notEqualToken, Value: "!="},
		},
		{
			name:  "end of line",
			input: "",
			want:  Token{Type: endOfStringToken, Value: ""},
		},
		{
			name:  "string literal",
			input: `"foo"`,
			want:  Token{Type: stringToken, Value: "foo"},
		},
		{
			name:  "boolean literal true",
			input: "true",
			want:  Token{Type: boolToken, Value: "true"},
		},
		{
			name:  "boolean literal false",
			input: "false",
			want:  Token{Type: boolToken, Value: "false"},
		},
		{
			name:  "matches operator",
			input: "matches",
			want:  Token{Type: matchesToken, Value: "matches"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &lexer{
				input: strings.NewReader(tt.input),
			}
			if got := l.Tokenize(); got != tt.want {
				t.Errorf("lexer.Tokenize() got = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestArrayTokens(t *testing.T) {
	input := "[a,  b,c]"
	lexer := newLexer(input)
	var token Token
	var tokens []Token
	for token.Type != endOfStringToken {
		token = lexer.Tokenize()
		if token.Type == errorToken {
			t.Fatal(token.Value)
		}
		tokens = append(tokens, token)
	}
	want := []Token{
		Token{Type: leftSquareToken, Value: "["},
		Token{Type: identifierToken, Value: "a"},
		Token{Type: commaToken, Value: ","},
		Token{Type: identifierToken, Value: "b"},
		Token{Type: commaToken, Value: ","},
		Token{Type: identifierToken, Value: "c"},
		Token{Type: rightSquareToken, Value: "]"},
		Token{Type: endOfStringToken},
	}

	if got := tokens; !reflect.DeepEqual(got, want) {
		t.Fatalf("lexer.Tokenize(): %v != %v", got, want)
	}
}

func TestLexIdentifiers(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Token
	}{
		{
			name: "good identifier 1",
			input: "	_0.a1 ",
			want: Token{Type: identifierToken, Value: "_0.a1"},
		},
		{
			name:  "good identifier 2",
			input: "__CamelCase__.snake_case",
			want:  Token{Type: identifierToken, Value: "__CamelCase__.snake_case"},
		},
		{
			name:  "bad identifier 1",
			input: "_0.1",
			want:  Token{Type: errorToken, Value: `invalid identifier: "_0.1"`},
		},
		{
			name:  "bad identifier 2",
			input: "0asdf",
			want:  Token{Type: errorToken, Value: `invalid rune: "0"`},
		},
		{
			name:  "bad identifier 3",
			input: "asdf.",
			want:  Token{Type: errorToken, Value: `end of input while scanning identifier: "asdf."`},
		},
		{
			name:  "bad identifier 4",
			input: ".asdf",
			want:  Token{Type: errorToken, Value: `invalid rune: "."`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &lexer{
				input: strings.NewReader(tt.input),
			}
			if got := l.Tokenize(); got != tt.want {
				t.Errorf("lexer.Tokenize() got = %v, want %v", got, tt.want)
			}
		})
	}
}
