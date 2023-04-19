package dynamic_test

import (
	"encoding/json"
	"testing"

	v2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/dynamic"
)

func BenchmarkSynthesize(b *testing.B) {
	c := v2.FixtureCheck("foo")
	for i := 0; i < b.N; i++ {
		_ = dynamic.Synthesize(c)
	}
}

func BenchmarkCheckMarshalRoundtrip(b *testing.B) {
	c := v2.FixtureCheck("foo")
	bytez, _ := json.Marshal(c)
	for i := 0; i < b.N; i++ {
		var check v2.Check
		_ = json.Unmarshal(bytez, &check)
	}
}
