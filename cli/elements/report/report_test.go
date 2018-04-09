package report

import (
	"fmt"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReportHasWarnings(t *testing.T) {
	testCases := []struct {
		entries []logrus.Level
		want    bool
	}{
		{[]logrus.Level{logrus.DebugLevel}, false},
		{[]logrus.Level{logrus.ErrorLevel, logrus.DebugLevel}, false},
		{[]logrus.Level{logrus.ErrorLevel, logrus.WarnLevel}, true},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("[%v] entries should return %t", tc.entries, tc.want), func(t *testing.T) {
			report := Report{}
			for _, e := range tc.entries {
				report.AddEntry(Entry{Level: e})
			}
			assert.Equal(t, tc.want, report.HasWarnings())
		})
	}
}

func TestReportHasErrors(t *testing.T) {
	testCases := []struct {
		entries []logrus.Level
		want    bool
	}{
		{[]logrus.Level{logrus.DebugLevel}, false},
		{[]logrus.Level{logrus.ErrorLevel, logrus.DebugLevel}, true},
		{[]logrus.Level{logrus.ErrorLevel, logrus.WarnLevel}, true},
		{[]logrus.Level{logrus.FatalLevel, logrus.WarnLevel}, true},
		{[]logrus.Level{logrus.DebugLevel, logrus.WarnLevel}, false},
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("[%v] entries should return %t", tc.entries, tc.want), func(t *testing.T) {
			report := Report{}
			for _, e := range tc.entries {
				report.AddEntry(Entry{Level: e})
			}
			assert.Equal(t, report.HasErrors(), tc.want)
		})
	}
}
func TestReportFlush(t *testing.T) {
	assert := assert.New(t)

	out := exWriter{}
	genReport := func(level logrus.Level) Report {
		r := New()
		r.LogLevel = level
		r.Out = &out
		r.entries = []Entry{
			{Level: logrus.DebugLevel, Message: "Test"},
			{Level: logrus.WarnLevel, Message: "Test"},
			{Level: logrus.ErrorLevel, Message: "Test"},
			{Level: logrus.FatalLevel, Message: "Test"},
		}
		return r
	}

	report := genReport(logrus.DebugLevel)
	assert.Len(report.entries, 4, "All configured entries are present in report")

	require.NoError(t, report.Flush())
	assert.Len(report.entries, 0, "All configured entries have been flushed from report")
	assert.NotEmpty(out.result, "Flush should have written to configured io.Writer")
	assert.Len(strings.Split(out.result, "\n"), 5, "All entries have been written to configured io.Writer")

	report = genReport(logrus.InfoLevel)
	assert.Len(report.entries, 4, "All configured entries are present in report")

	out.Clean()
	require.NoError(t, report.Flush())
	assert.Len(strings.Split(out.result, "\n"), 4, "All entries have been written to configured io.Writer except debug entry")
}

type exWriter struct {
	result string
}

func (w *exWriter) Clean() {
	w.result = ""
}

func (w *exWriter) Write(p []byte) (int, error) {
	w.result += string(p)
	return 0, nil
}
