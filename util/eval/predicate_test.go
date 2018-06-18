package eval

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvaluatePredicate(t *testing.T) {
	type args struct {
		expression string
		parameters map[string]interface{}
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
		errType string
	}{
		{
			name: "unparsable expression",
			args: args{
				expression: "1 &&",
			},
			wantErr: true,
			errType: "SyntaxError",
		},
		{
			name: "unevaluable expression",
			args: args{
				expression: "foo > 0",
			},
			wantErr: true,
		},
		{
			name: "non-boolean expression value",
			args: args{
				expression: "42",
			},
			wantErr: true,
			errType: "TypeError",
		},
		{
			name: "positive result",
			args: args{
				expression: "foo > 1",
				parameters: map[string]interface{}{
					"foo": 2,
				},
			},
			want: true,
		},
		{
			name: "negative result",
			args: args{
				expression: "foo > 1",
				parameters: map[string]interface{}{
					"foo": 0,
				},
			},
			want: false,
		},
		{
			name: "negative hour",
			args: args{
				expression: "hour(timestamp) == 19",
				parameters: map[string]interface{}{
					"timestamp": 1520275913, // Monday, March 5, 2018 6:51:53 PM UTC
				},
			},
			want: false,
		},
		{
			name: "positive hour",
			args: args{
				expression: "hour(timestamp) == 18",
				parameters: map[string]interface{}{
					"timestamp": 1520275913, // Monday, March 5, 2018 6:51:53 PM UTC
				},
			},
			want: true,
		},
		{
			name: "positive between hour",
			args: args{
				expression: "hour(timestamp) >= 17 && hour(timestamp) <= 19",
				parameters: map[string]interface{}{
					"timestamp": 1520275913, // Monday, March 5, 2018 6:51:53 PM UTC
				},
			},
			want: true,
		},
		{
			name: "positive weekday",
			args: args{
				expression: "weekday(timestamp) == 1",
				parameters: map[string]interface{}{
					"timestamp": 1520275913, // Monday, March 5, 2018 6:51:53 PM UTC
				},
			},
			want: true,
		},
		{
			name: "negative weekday",
			args: args{
				expression: "weekday(timestamp) == 2",
				parameters: map[string]interface{}{
					"timestamp": 1520275913, // Monday, March 5, 2018 6:51:53 PM UTC
				},
			},
			want: false,
		},
		{
			name: "positive weekday array",
			args: args{
				expression: "weekday(timestamp) in (1, 2, 3, 4, 5)",
				parameters: map[string]interface{}{
					"timestamp": 1520275913,
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EvaluatePredicate(tt.args.expression, tt.args.parameters)
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil && tt.wantErr && tt.errType != "" {
				switch tt.errType {
				case "SyntaxError":
					if _, ok := err.(SyntaxError); !ok {
						t.Error("want SyntaxError")
					}
				case "TypeError":
					if _, ok := err.(TypeError); !ok {
						t.Error("want TypeError")
					}
				}
			}
			if got != tt.want {
				t.Errorf("Evaluate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateStatements(t *testing.T) {
	// Valid statement
	statements := []string{"10 > 0"}
	assert.NoError(t, ValidateStatements(statements, false))

	// Invalid statement
	statements = []string{"10. 0"}
	assert.Error(t, ValidateStatements(statements, false))

	// Forbidden modifier token
	statements = []string{"10 + 2 > 0"}
	assert.Error(t, ValidateStatements(statements, true))

	// Allowed modifier token
	statements = []string{"10 + 2 > 0"}
	assert.NoError(t, ValidateStatements(statements, false))
}
