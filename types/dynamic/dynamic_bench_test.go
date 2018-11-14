package dynamic_test

import (
	"encoding/json"
	"testing"

	jsoniter "github.com/json-iterator/go"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/types/dynamic"
)

func BenchmarkSynthesize(b *testing.B) {
	c := types.FixtureCheck("foo")
	for i := 0; i < b.N; i++ {
		_ = dynamic.Synthesize(c)
	}
}

func BenchmarkCheckMarshalRoundtrip(b *testing.B) {
	c := types.FixtureCheck("foo")
	bytez, _ := json.Marshal(c)
	for i := 0; i < b.N; i++ {
		var check types.Check
		_ = jsoniter.Unmarshal(bytez, &check)
	}
}
