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
	}{
		{
			name: "unparsable expression",
			args: args{
				expression: "1 &&",
			},
			wantErr: true,
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := EvaluatePredicate(tt.args.expression, tt.args.parameters)
			if (err != nil) != tt.wantErr {
				t.Errorf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
				return
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
