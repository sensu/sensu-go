package v2

import (
	"encoding/json"
	"testing"
)

func BenchmarkCheckRequestMarshal(b *testing.B) {
	req := FixtureCheckRequest("cake")
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = json.Marshal(req)
		}
	})
}

func BenchmarkCheckRequestUnmarshal(b *testing.B) {
	req := FixtureCheckRequest("cake")
	data, err := json.Marshal(req)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var req CheckRequest
			_ = json.Unmarshal(data, &req)
		}
	})
}

func BenchmarkCheckConfigMarshal(b *testing.B) {
	req := FixtureCheckConfig("cake")
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = json.Marshal(req)
		}
	})
}

func BenchmarkCheckConfigUnmarshal(b *testing.B) {
	req := FixtureCheckConfig("cake")
	data, err := json.Marshal(req)
	if err != nil {
		b.Fatal(err)
	}
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var req CheckConfig
			_ = json.Unmarshal(data, &req)
		}
	})
}
