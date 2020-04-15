package logging

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestIncrementLogLevel(t *testing.T) {
	tests := []struct {
		Level    logrus.Level
		ExpLevel logrus.Level
	}{
		{
			Level:    logrus.ErrorLevel,
			ExpLevel: logrus.WarnLevel,
		},
		{
			Level:    logrus.WarnLevel,
			ExpLevel: logrus.InfoLevel,
		},
		{
			Level:    logrus.InfoLevel,
			ExpLevel: logrus.DebugLevel,
		},
		{
			Level:    logrus.DebugLevel,
			ExpLevel: logrus.TraceLevel,
		},
		{
			Level:    logrus.TraceLevel,
			ExpLevel: logrus.ErrorLevel,
		},
	}
	for _, test := range tests {
		t.Run(test.Level.String(), func(t *testing.T) {
			if got, want := IncrementLogLevel(test.Level), test.ExpLevel; got != want {
				t.Fatalf("bad log level: got %s, want %s", got, want)
			}
		})
	}
}
