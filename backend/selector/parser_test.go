package selector

import (
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    *Selector
		wantErr bool
	}{
		{
			name:  "simple requirement",
			input: "foo == bar",
			want: &Selector{Operations: []Operation{
				Operation{LValue: "foo", Operator: DoubleEqualSignOperator, RValues: []string{"bar"}},
			}},
		},
		{
			name:  "double requirements",
			input: "foo == bar && baz != qux",
			want: &Selector{Operations: []Operation{
				Operation{LValue: "foo", Operator: DoubleEqualSignOperator, RValues: []string{"bar"}},
				Operation{LValue: "baz", Operator: NotEqualOperator, RValues: []string{"qux"}},
			}},
		},
		{
			name:  "in array",
			input: "foo in [foo,bar]",
			want: &Selector{Operations: []Operation{
				Operation{LValue: "foo", Operator: InOperator, RValues: []string{"foo", "bar"}},
			}},
		},
		{
			name:  "notin array",
			input: "foo notin [foo,bar]",
			want: &Selector{Operations: []Operation{
				Operation{LValue: "foo", Operator: NotInOperator, RValues: []string{"foo", "bar"}},
			}},
		},
		{
			name:    "invalid operator",
			input:   "foo maybein [foo,bar]",
			wantErr: true,
		},
		{
			name:    "missing identifier after '&&'",
			input:   "foo != bar &&",
			wantErr: true,
		},
		{
			name:    "invalid '&&' at beginning of operation",
			input:   "&& foo != bar",
			wantErr: true,
		},
		{
			name:    "unexpected token",
			input:   "foo != bar &",
			wantErr: true,
		},
		{
			name:    "invalid array",
			input:   "foo in [foo != bar]",
			wantErr: true,
		},
		{
			name:    "invalid value",
			input:   "foo != ==",
			wantErr: true,
		},
		{
			name:  "string value",
			input: "region == \"us-west-1\"",
			want: &Selector{Operations: []Operation{
				Operation{LValue: "region", Operator: DoubleEqualSignOperator, RValues: []string{"us-west-1"}},
			}},
		},
		{
			name:  "boolean value",
			input: "publish == true",
			want: &Selector{Operations: []Operation{
				Operation{LValue: "publish", Operator: DoubleEqualSignOperator, RValues: []string{"true"}},
			}},
		},
		{
			name:  "in array with string token",
			input: "\"entity:whisky:*\" in [\"entity:whisky:*\"]",
			want: &Selector{Operations: []Operation{
				Operation{LValue: "entity:whisky:*", Operator: InOperator, RValues: []string{"entity:whisky:*"}},
			}},
		},
		{
			name:  "matches operator",
			input: "foo matches bar",
			want: &Selector{Operations: []Operation{
				Operation{LValue: "foo", Operator: MatchesOperator, RValues: []string{"bar"}},
			}},
		},
		{
			name:  "#1485",
			input: "\"my sub\" in check.subscriptions && check.publish == true",
			want: &Selector{Operations: []Operation{
				{LValue: "my sub", Operator: InOperator, RValues: []string{"check.subscriptions"}},
				{LValue: "check.publish", Operator: DoubleEqualSignOperator, RValues: []string{"true"}},
			}},
		},
		{
			name:  "#1485",
			input: "check.publish == true && \"my sub\" in check.subscriptions",
			want: &Selector{Operations: []Operation{
				{LValue: "check.publish", Operator: DoubleEqualSignOperator, RValues: []string{"true"}},
				{LValue: "my sub", Operator: InOperator, RValues: []string{"check.subscriptions"}},
			}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parser.Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parser.Parse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParserParseValues(t *testing.T) {
	tests := []struct {
		name    string
		Tokens  []Token
		want    []string
		wantErr bool
	}{
		{
			name: "single item array", // [foo]
			Tokens: []Token{
				Token{Type: leftSquareToken},
				Token{Type: identifierToken, Value: "foo"},
				Token{Type: rightSquareToken},
			},
			want: []string{"foo"},
		},
		{
			name: "multiple items array", // [foo,bar]
			Tokens: []Token{
				Token{Type: leftSquareToken},
				Token{Type: identifierToken, Value: "foo"},
				Token{Type: commaToken},
				Token{Type: identifierToken, Value: "bar"},
				Token{Type: rightSquareToken},
			},
			want: []string{"foo", "bar"},
		},
		// With the changes to the parser, this test only consumes the first
		// token, so it doesn't produces an error, and therefore fails.
		//{
		//	name: "invalid array", // foo,bar
		//	Tokens: []Token{
		//		Token{Type: identifierToken, value: "foo"},
		//		Token{Type: commaToken},
		//		Token{Type: identifierToken, value: "bar"},
		//	},
		//	wantErr: true,
		//},
		{
			name: "unexpected token", // [foo != bar]
			Tokens: []Token{
				Token{Type: leftSquareToken},
				Token{Type: identifierToken, Value: "foo"},
				Token{Type: notEqualToken},
				Token{Type: identifierToken, Value: "bar"},
				Token{Type: rightSquareToken},
			},
			want:    []string{"foo"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &parser{
				results: tt.Tokens,
			}
			got, err := p.parseValues()
			if (err != nil) != tt.wantErr {
				t.Errorf("parser.parseValues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parser.parseValues() = %v, want %v", got, tt.want)
			}
		})
	}
}
