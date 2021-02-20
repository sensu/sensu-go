package etcd

import (
	"testing"

	"github.com/stretchr/testify/mock"
	zapcore "go.uber.org/zap/zapcore"
)

// MockPrimitiveArrayEncoder is a mock zapcore.PrimitiveArrayEncoder used for
// testing purposes.
type MockPrimitiveArrayEncoder struct {
	mock.Mock
}

func (m MockPrimitiveArrayEncoder) AppendBool(v bool)             { m.Called(v) }
func (m MockPrimitiveArrayEncoder) AppendByteString(v []byte)     { m.Called(v) }
func (m MockPrimitiveArrayEncoder) AppendComplex64(v complex64)   { m.Called(v) }
func (m MockPrimitiveArrayEncoder) AppendComplex128(v complex128) { m.Called(v) }
func (m MockPrimitiveArrayEncoder) AppendFloat32(v float32)       { m.Called(v) }
func (m MockPrimitiveArrayEncoder) AppendFloat64(v float64)       { m.Called(v) }
func (m MockPrimitiveArrayEncoder) AppendInt(v int)               { m.Called(v) }
func (m MockPrimitiveArrayEncoder) AppendInt8(v int8)             { m.Called(v) }
func (m MockPrimitiveArrayEncoder) AppendInt16(v int16)           { m.Called(v) }
func (m MockPrimitiveArrayEncoder) AppendInt32(v int32)           { m.Called(v) }
func (m MockPrimitiveArrayEncoder) AppendInt64(v int64)           { m.Called(v) }
func (m MockPrimitiveArrayEncoder) AppendString(v string)         { m.Called(v) }
func (m MockPrimitiveArrayEncoder) AppendUint(v uint)             { m.Called(v) }
func (m MockPrimitiveArrayEncoder) AppendUint8(v uint8)           { m.Called(v) }
func (m MockPrimitiveArrayEncoder) AppendUint16(v uint16)         { m.Called(v) }
func (m MockPrimitiveArrayEncoder) AppendUint32(v uint32)         { m.Called(v) }
func (m MockPrimitiveArrayEncoder) AppendUint64(v uint64)         { m.Called(v) }
func (m MockPrimitiveArrayEncoder) AppendUintptr(v uintptr)       { m.Called(v) }

func Test_sensuLevelEncoder(t *testing.T) {
	type args struct {
		l zapcore.Level
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "zapcore.InfoLevel should return info",
			args: args{
				l: zapcore.InfoLevel,
			},
			want: "info",
		},
		{
			name: "zapcore.WarnLevel should return warning",
			args: args{
				l: zapcore.WarnLevel,
			},
			want: "warning",
		},
		{
			name: "zapcore.ErrorLevel should return error",
			args: args{
				l: zapcore.ErrorLevel,
			},
			want: "error",
		},
		{
			name: "zapcore.PanicLevel should return panic",
			args: args{
				l: zapcore.PanicLevel,
			},
			want: "panic",
		},
		{
			name: "zapcore.FatalLevel should return fatal",
			args: args{
				l: zapcore.FatalLevel,
			},
			want: "fatal",
		},
		{
			name: "zapcore.DebugLevel should return debug",
			args: args{
				l: zapcore.DebugLevel,
			},
			want: "debug",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string

			enc := MockPrimitiveArrayEncoder{}
			enc.On("AppendString", mock.MatchedBy(func(s string) bool {
				got = s
				return true
			})).Return()

			sensuLevelEncoder(tt.args.l, enc)

			if got != tt.want {
				t.Errorf("sensuLevelEncoder() got = %s, want %s", got, tt.want)
			}
		})
	}
}
