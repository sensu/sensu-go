package v2

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/sensu/sensu-go/types/dynamic"
)

var processes []*Process

func init() {
	processes = FixtureProcesses(500)
}

func BenchmarkEntitySynthesizingWithoutProcesses(b *testing.B) {
	for n := 0; n < b.N; n++ {
		_ = dynamic.Synthesize(FixtureEntity(fmt.Sprintf("entity-%d", n)))
	}
}

func BenchmarkEntitySynthesizingWithProcesses(b *testing.B) {
	for n := 0; n < b.N; n++ {
		entity := FixtureEntity(fmt.Sprintf("entity-%d", n))
		entity.System.Processes = processes
		_ = dynamic.Synthesize(entity)
	}
}

func BenchmarkEntityUnmarshalWithoutProcesses(b *testing.B) {
	entity := FixtureEntity("foo")
	bytes, err := json.Marshal(entity)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		var e Entity
		_ = json.Unmarshal(bytes, &e)
	}
}

func BenchmarkEntityUnmarshalWithProcesses(b *testing.B) {
	entity := FixtureEntity("foo")
	entity.System.Processes = processes
	bytes, err := json.Marshal(entity)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		var e Entity
		_ = json.Unmarshal(bytes, &e)
	}
}
