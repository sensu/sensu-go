package statsd

import (
	"testing"

	"github.com/atlassian/gostatsd"
	"github.com/atlassian/gostatsd/pkg/pool"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsLexer(t *testing.T) {
	t.Parallel()
	tests := map[string]gostatsd.Metric{
		"foo.bar.baz:2|c":               {Name: "foo.bar.baz", Value: 2, Type: gostatsd.COUNTER},
		"abc.def.g:3|g":                 {Name: "abc.def.g", Value: 3, Type: gostatsd.GAUGE},
		"def.g:10|ms":                   {Name: "def.g", Value: 10, Type: gostatsd.TIMER},
		"def.h:10|h":                    {Name: "def.h", Value: 10, Type: gostatsd.TIMER},
		"def.i:10|h|#foo":               {Name: "def.i", Value: 10, Type: gostatsd.TIMER, Tags: gostatsd.Tags{"foo"}},
		"smp.rte:5|c|@0.1":              {Name: "smp.rte", Value: 50, Type: gostatsd.COUNTER},
		"smp.rte:5|c|@0.1|#foo:bar,baz": {Name: "smp.rte", Value: 50, Type: gostatsd.COUNTER, Tags: gostatsd.Tags{"foo:bar", "baz"}},
		"smp.rte:5|c|#foo:bar,baz":      {Name: "smp.rte", Value: 5, Type: gostatsd.COUNTER, Tags: gostatsd.Tags{"foo:bar", "baz"}},
		"uniq.usr:joe|s":                {Name: "uniq.usr", StringValue: "joe", Type: gostatsd.SET},
		"fooBarBaz:2|c":                 {Name: "fooBarBaz", Value: 2, Type: gostatsd.COUNTER},
		"smp.rte:5|c|#Foo:Bar,baz":      {Name: "smp.rte", Value: 5, Type: gostatsd.COUNTER, Tags: gostatsd.Tags{"Foo:Bar", "baz"}},
		"smp.gge:1|g|#Foo:Bar":          {Name: "smp.gge", Value: 1, Type: gostatsd.GAUGE, Tags: gostatsd.Tags{"Foo:Bar"}},
		"smp.gge:1|g|#fo_o:ba-r":        {Name: "smp.gge", Value: 1, Type: gostatsd.GAUGE, Tags: gostatsd.Tags{"fo_o:ba-r"}},
		"smp gge:1|g":                   {Name: "smp_gge", Value: 1, Type: gostatsd.GAUGE},
		"smp/gge:1|g":                   {Name: "smp-gge", Value: 1, Type: gostatsd.GAUGE},
		"smp,gge$:1|g":                  {Name: "smpgge", Value: 1, Type: gostatsd.GAUGE},
		"un1qu3:john|s":                 {Name: "un1qu3", StringValue: "john", Type: gostatsd.SET},
		"un1qu3:john|s|#some:42":        {Name: "un1qu3", StringValue: "john", Type: gostatsd.SET, Tags: gostatsd.Tags{"some:42"}},
		"da-sh:1|s":                     {Name: "da-sh", StringValue: "1", Type: gostatsd.SET},
		"under_score:1|s":               {Name: "under_score", StringValue: "1", Type: gostatsd.SET},
		"a:1|g|#f,,":                    {Name: "a", Value: 1, Type: gostatsd.GAUGE, Tags: gostatsd.Tags{"f"}},
		"a:1|g|#,,f":                    {Name: "a", Value: 1, Type: gostatsd.GAUGE, Tags: gostatsd.Tags{"f"}},
		"a:1|g|#f,,z":                   {Name: "a", Value: 1, Type: gostatsd.GAUGE, Tags: gostatsd.Tags{"f", "z"}},
		"a:1|g|#":                       {Name: "a", Value: 1, Type: gostatsd.GAUGE},
		"a:1|g|#,":                      {Name: "a", Value: 1, Type: gostatsd.GAUGE},
		"a:1|g|#,,":                     {Name: "a", Value: 1, Type: gostatsd.GAUGE},
	}

	compareMetric(t, tests, "")
}

func TestInvalidMetricsLexer(t *testing.T) {
	t.Parallel()
	failing := []string{"fOO|bar:bazkk", "foo.bar.baz:1|q", "NaN.should.be:NaN|g"}
	for _, tc := range failing {
		tc := tc
		t.Run(tc, func(t *testing.T) {
			t.Parallel()
			result, _, err := parseLine([]byte(tc), "")
			assert.Error(t, err, result)
		})
	}

	tests := map[string]gostatsd.Metric{
		"foo.bar.baz:2|c": {Name: "stats.foo.bar.baz", Value: 2, Type: gostatsd.COUNTER},
		"abc.def.g:3|g":   {Name: "stats.abc.def.g", Value: 3, Type: gostatsd.GAUGE},
		"def.g:10|ms":     {Name: "stats.def.g", Value: 10, Type: gostatsd.TIMER},
		"uniq.usr:joe|s":  {Name: "stats.uniq.usr", StringValue: "joe", Type: gostatsd.SET},
	}

	compareMetric(t, tests, "stats")
}

func TestEventsLexer(t *testing.T) {
	t.Parallel()
	//_e{title.length,text.length}:title|text|d:date_happened|h:hostname|p:priority|t:alert_type|#tag1,tag2
	tests := map[string]gostatsd.Event{
		"_e{1,1}:a|b":                                               {Title: "a", Text: "b"},
		"_e{6,18}:ab|| c|hello,\\nmy friend!":                       {Title: "ab|| c", Text: "hello,\nmy friend!"},
		"_e{1,1}:a|b|d:123123":                                      {Title: "a", Text: "b", DateHappened: 123123},
		"_e{1,1}:a|b|d:123123|h:hoost":                              {Title: "a", Text: "b", DateHappened: 123123, Hostname: "hoost"},
		"_e{1,1}:a|b|d:123123|h:hoost|p:low":                        {Title: "a", Text: "b", DateHappened: 123123, Hostname: "hoost", Priority: gostatsd.PriLow},
		"_e{1,1}:a|b|d:123123|h:hoost|p:low|t:warning":              {Title: "a", Text: "b", DateHappened: 123123, Hostname: "hoost", Priority: gostatsd.PriLow, AlertType: gostatsd.AlertWarning},
		"_e{1,1}:a|b|d:123123|h:hoost|p:low|t:warning|#tag1,t:tag2": {Title: "a", Text: "b", DateHappened: 123123, Hostname: "hoost", Priority: gostatsd.PriLow, AlertType: gostatsd.AlertWarning, Tags: []string{"tag1", "t:tag2"}},
		"_e{1,1}:a|b|t:warning|d:123123|h:hoost|p:low|#tag1,t:tag2": {Title: "a", Text: "b", DateHappened: 123123, Hostname: "hoost", Priority: gostatsd.PriLow, AlertType: gostatsd.AlertWarning, Tags: []string{"tag1", "t:tag2"}},
		"_e{1,1}:a|b|p:low|t:warning|d:123123|h:hoost|#tag1,t:tag2": {Title: "a", Text: "b", DateHappened: 123123, Hostname: "hoost", Priority: gostatsd.PriLow, AlertType: gostatsd.AlertWarning, Tags: []string{"tag1", "t:tag2"}},
		"_e{1,1}:a|b|h:hoost|p:low|t:warning|d:123123|#tag1,t:tag2": {Title: "a", Text: "b", DateHappened: 123123, Hostname: "hoost", Priority: gostatsd.PriLow, AlertType: gostatsd.AlertWarning, Tags: []string{"tag1", "t:tag2"}},

		"_e{1,1}:a|b|h:hoost":      {Title: "a", Text: "b", Hostname: "hoost"},
		"_e{1,1}:a|b|p:low":        {Title: "a", Text: "b", Priority: gostatsd.PriLow},
		"_e{1,1}:a|b|t:warning":    {Title: "a", Text: "b", AlertType: gostatsd.AlertWarning},
		"_e{1,1}:a|b|#tag1,t:tag2": {Title: "a", Text: "b", Tags: []string{"tag1", "t:tag2"}},
		"_e{20,34}:Deployment completed|Deployment completed in 7 minutes.|d:1463746133|h:9c00cf070c14|s:Micros Server|t:success|#topic:service.deploy,message_env:pdev,service_id:node-refapp-ci-internal,deployment_id:72e95b0f-37b0-4cf9-8e92-3e47d006b63f": {
			Title:          "Deployment completed",
			Text:           "Deployment completed in 7 minutes.",
			DateHappened:   1463746133,
			Hostname:       "9c00cf070c14",
			SourceTypeName: "Micros Server",
			Tags:           []string{"topic:service.deploy", "message_env:pdev", "service_id:node-refapp-ci-internal", "deployment_id:72e95b0f-37b0-4cf9-8e92-3e47d006b63f"},
			AlertType:      gostatsd.AlertSuccess,
		},
	}

	for input, expected := range tests {
		input := input
		expected := expected
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			_, result, err := parseLine([]byte(input), "")
			require.NoError(t, err)
			assert.Equal(t, &expected, result)
		})
	}
}

func TestInvalidEventsLexer(t *testing.T) {
	t.Parallel()
	failing := map[string]error{
		"_x{1,1}:a|b": errInvalidType,
		"_e{2,1}:a|b": errNotEnoughData,
		"_e{1,2}:a|b": errNotEnoughData,
		"_e{2,2}:a|b": errNotEnoughData,
		"_e{1,1}:ab":  errNotEnoughData,
		"_e{1,1}ab":   errInvalidFormat,
		"_e{1,1}a:b":  errInvalidFormat,
		"_e{1,1}:a:b": errInvalidFormat,
		"_e{,1}:a|b":  errInvalidFormat,
		"_e{1,}:a|b":  errInvalidFormat,
		"_e{1}:a|b":   errInvalidFormat,
		"_e{}:a|b":    errInvalidFormat,
		"_e1,2}:a|b":  errInvalidFormat,
		"_e:a|b":      errInvalidFormat,
		"_e{999999999999999999999999,1}:a|b": errOverflow,
		"_e{1,999999999999999999999999}:a|b": errOverflow,
	}
	for input, expectedErr := range failing {
		input := input
		expectedErr := expectedErr
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			m, e, err := parseLine([]byte(input), "")
			assert.Equal(t, expectedErr, err)
			assert.Nil(t, m)
			assert.Nil(t, e)
		})
	}
}

func parseLine(input []byte, namespace string) (*gostatsd.Metric, *gostatsd.Event, error) {
	l := lexer{
		metricPool: pool.NewMetricPool(0),
	}
	return l.run(input, namespace)
}

func compareMetric(t *testing.T, tests map[string]gostatsd.Metric, namespace string) {
	for input, expected := range tests {
		input := input
		expected := expected
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			result, _, err := parseLine([]byte(input), namespace)
			result.DoneFunc = nil // Clear DoneFunc because it contains non-predictable variable data which interferes with the tests
			require.NoError(t, err)
			assert.Equal(t, &expected, result)
		})
	}
}

var parselineBlackhole *gostatsd.Metric

func benchmarkLexer(dp *DatagramParser, input string, b *testing.B) {
	slice := []byte(input)
	var r *gostatsd.Metric
	dp.metricPool = pool.NewMetricPool(0)
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		r, _, _ = dp.parseLine(slice)
		r.Done()
	}
	parselineBlackhole = r
}

func BenchmarkParseCounter(b *testing.B) {
	benchmarkLexer(&DatagramParser{}, "foo.bar.baz:2|c", b)
}
func BenchmarkParseCounterWithSampleRate(b *testing.B) {
	benchmarkLexer(&DatagramParser{}, "smp.rte:5|c|@0.1", b)
}
func BenchmarkParseCounterWithTags(b *testing.B) {
	benchmarkLexer(&DatagramParser{}, "smp.rte:5|c|#foo:bar,baz", b)
}
func BenchmarkParseCounterWithTagsAndSampleRate(b *testing.B) {
	benchmarkLexer(&DatagramParser{}, "smp.rte:5|c|@0.1|#foo:bar,baz", b)
}
func BenchmarkParseGauge(b *testing.B) {
	benchmarkLexer(&DatagramParser{}, "abc.def.g:3|g", b)
}
func BenchmarkParseTimer(b *testing.B) {
	benchmarkLexer(&DatagramParser{}, "def.g:10|ms", b)
}
func BenchmarkParseSet(b *testing.B) {
	benchmarkLexer(&DatagramParser{}, "uniq.usr:joe|s", b)
}
func BenchmarkParseCounterWithDefaultTags(b *testing.B) {
	benchmarkLexer(&DatagramParser{}, "foo.bar.baz:2|c", b)
}
func BenchmarkParseCounterWithDefaultTagsAndTags(b *testing.B) {
	benchmarkLexer(&DatagramParser{}, "foo.bar.baz:2|c|#foo:bar,baz", b)
}
func BenchmarkParseCounterWithDefaultTagsAndTagsAndNameSpace(b *testing.B) {
	benchmarkLexer(&DatagramParser{namespace: "stats"}, "foo.bar.baz:2|c|#foo:bar,baz", b)
}
