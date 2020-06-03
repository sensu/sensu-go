package dynamic_test

import (
	"encoding/json"
	"testing"

	jsoniter "github.com/json-iterator/go"
	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/api/core/v2/internal/types/dynamic"
)

func BenchmarkSynthesize(b *testing.B) {
	c := corev2.FixtureCheck("foo")
	for i := 0; i < b.N; i++ {
		_ = dynamic.Synthesize(c)
	}
}

func BenchmarkCheckMarshalRoundtrip(b *testing.B) {
	c := corev2.FixtureCheck("foo")
	bytez, _ := json.Marshal(c)
	for i := 0; i < b.N; i++ {
		var check corev2.Check
		_ = jsoniter.Unmarshal(bytez, &check)
	}
}
