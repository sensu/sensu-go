package js_test

import (
	"testing"

	"github.com/sensu/sensu-go/js"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-go/types/dynamic"
)

func BenchmarkCheckEval(b *testing.B) {
	check := types.FixtureCheck("foo")
	for i := 0; i < b.N; i++ {
		synth := dynamic.Synthesize(check)
		params := map[string]interface{}{
			"check": synth,
		}
		_, _ = js.Evaluate("check.status == 0", params, nil)
	}
}
