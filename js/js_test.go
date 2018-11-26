package js

import "testing"

func BenchmarkNewOttoVM(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = newOttoVM(nil)
		for k := range ottoCache.vms {
			delete(ottoCache.vms, k)
		}
	}
}
